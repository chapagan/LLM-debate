package ws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"llmdebate/internal/debate"
	"llmdebate/internal/protocol"
)

type Client struct {
	conn                   *websocket.Conn
	send                   chan protocol.OutboundEvent
	sessionCancel          context.CancelFunc
	mu                     sync.Mutex
	sendMu                 sync.Mutex
	closed                 bool
	lastDebateStart        time.Time
	debateStartMinInterval time.Duration
}

const defaultDebateStartMinInterval = 10 * time.Second

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn:                   conn,
		send:                   make(chan protocol.OutboundEvent, 32),
		debateStartMinInterval: defaultDebateStartMinInterval,
	}
}

func (c *Client) tryStartDebate(now time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	interval := c.debateStartMinInterval
	if interval <= 0 {
		interval = defaultDebateStartMinInterval
	}
	if !c.lastDebateStart.IsZero() && now.Sub(c.lastDebateStart) < interval {
		return errDebateRateLimited
	}
	c.lastDebateStart = now
	return nil
}

var errDebateRateLimited = errors.New("debate start rate limited; wait before starting another debate")

func (c *Client) setSession(cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sessionCancel != nil {
		c.sessionCancel()
	}
	c.sessionCancel = cancel
}

func (c *Client) cancelSession() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sessionCancel != nil {
		c.sessionCancel()
		c.sessionCancel = nil
	}
}

type Hub struct {
	mu      sync.Mutex
	clients map[*Client]struct{}
	limiter *DebateStartLimiter
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		limiter: NewDebateStartLimiter(DebateStartLimiterConfig{}),
	}
}

func (h *Hub) allowDebateStart(ip string, now time.Time) error {
	if h == nil || h.limiter == nil {
		return nil
	}
	return h.limiter.Allow(ip, now)
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	c.cancelSession()
	c.closeSend()
}

type Handler struct {
	Hub    *Hub
	Runner debate.Runner
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Hub == nil {
		http.Error(w, "websocket hub is not configured", http.StatusInternalServerError)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	defer conn.Close(websocket.StatusNormalClosure, "")
	client := NewClient(conn)
	h.Hub.register(client)
	defer h.Hub.unregister(client)
	go client.writeLoop(ctx, cancel)
	client.readLoop(ctx, h.Runner, h.Hub, clientIP(r.RemoteAddr))
}

func (c *Client) readLoop(ctx context.Context, runner debate.Runner, hub *Hub, ip string) {
	defer c.cancelSession()
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
		var inbound protocol.InboundEvent
		if err := json.Unmarshal(data, &inbound); err != nil {
			c.enqueue(protocol.OutboundEvent{Type: protocol.EventError, Message: "invalid JSON"})
			continue
		}
		if inbound.Action != protocol.ActionStartDebate {
			c.enqueue(protocol.OutboundEvent{Type: protocol.EventError, Message: "unknown action"})
			continue
		}
		now := time.Now()
		if err := hub.allowDebateStart(ip, now); err != nil {
			c.enqueue(protocol.OutboundEvent{Type: protocol.EventError, Message: err.Error()})
			continue
		}
		if err := c.tryStartDebate(now); err != nil {
			c.enqueue(protocol.OutboundEvent{Type: protocol.EventError, Message: err.Error()})
			continue
		}
		sessionCtx, cancel := context.WithCancel(ctx)
		c.setSession(cancel)
		go func(topic string) {
			if err := runner.Run(sessionCtx, topic, c.enqueue); err != nil && sessionCtx.Err() == nil {
				c.enqueue(protocol.OutboundEvent{Type: protocol.EventError, Message: err.Error()})
			}
		}(inbound.Topic)
	}
}

func (c *Client) enqueue(event protocol.OutboundEvent) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.closed {
		return context.Canceled
	}
	select {
	case c.send <- event:
		return nil
	default:
		return context.DeadlineExceeded
	}
}

func (c *Client) closeSend() {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	close(c.send)
}

func (c *Client) writeLoop(ctx context.Context, cancelConnection context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			data, err := json.Marshal(event)
			if err == nil {
				err = c.conn.Write(writeCtx, websocket.MessageText, data)
			}
			cancel()
			if err != nil {
				c.cancelSession()
				cancelConnection()
				c.conn.Close(websocket.StatusInternalError, "write failed")
				return
			}
		}
	}
}

func encodeEvent(w io.Writer, event protocol.OutboundEvent) error {
	return json.NewEncoder(w).Encode(event)
}
