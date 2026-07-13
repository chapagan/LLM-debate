package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"llmdebate/internal/protocol"
)

func TestClientRateLimitsDebateStarts(t *testing.T) {
	client := NewClient(nil)
	client.debateStartMinInterval = 50 * time.Millisecond

	if err := client.tryStartDebate(time.Now()); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if err := client.tryStartDebate(time.Now()); err == nil {
		t.Fatal("expected rate limit on immediate second start")
	}

	if err := client.tryStartDebate(time.Now().Add(60 * time.Millisecond)); err != nil {
		t.Fatalf("start after interval: %v", err)
	}
}

func TestClientCancelsPreviousSession(t *testing.T) {
	client := NewClient(nil)
	ctx := context.Background()
	first, firstCancel := context.WithCancel(ctx)
	client.setSession(firstCancel)
	second, secondCancel := context.WithCancel(ctx)
	client.setSession(secondCancel)
	select {
	case <-first.Done():
	default:
		t.Fatal("first session was not canceled")
	}
	select {
	case <-second.Done():
		t.Fatal("second session should still be active")
	default:
	}
	client.cancelSession()
	select {
	case <-second.Done():
	default:
		t.Fatal("second session was not canceled")
	}
}

func TestEncodeOutboundEvent(t *testing.T) {
	var buf bytes.Buffer
	err := encodeEvent(&buf, protocol.OutboundEvent{Type: protocol.EventError, Message: "bad topic"})
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	var got protocol.OutboundEvent
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if got.Type != protocol.EventError || got.Message != "bad topic" {
		t.Fatalf("event = %+v", got)
	}
}

func TestHandlerReturnsServerErrorWhenHubIsNil(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	Handler{}.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandlerEmitsErrorForMalformedJSON(t *testing.T) {
	server := httptest.NewServer(Handler{Hub: NewHub()})
	defer server.Close()

	dialURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), dialURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := conn.Write(ctx, websocket.MessageText, []byte("{")); err != nil {
		t.Fatalf("write malformed JSON: %v", err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read error event: %v", err)
	}
	var got protocol.OutboundEvent
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if got.Type != protocol.EventError || got.Message != "invalid JSON" {
		t.Fatalf("event = %+v", got)
	}
}

func TestWriteLoopCancelsConnectionContextOnWriteFailure(t *testing.T) {
	triggerWrite := make(chan struct{})
	accepted := make(chan struct{})
	canceled := make(chan struct{})
	writeDone := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"127.0.0.1:*"},
		})
		if err != nil {
			t.Errorf("accept websocket: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		client := NewClient(conn)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			client.writeLoop(ctx, cancel)
			close(writeDone)
		}()

		close(accepted)
		<-triggerWrite
		if err := conn.CloseNow(); err != nil {
			t.Errorf("close server websocket: %v", err)
			return
		}
		if err := client.enqueue(protocol.OutboundEvent{Type: protocol.EventError, Message: "write failure"}); err != nil {
			t.Errorf("enqueue event: %v", err)
			return
		}

		select {
		case <-ctx.Done():
			close(canceled)
		case <-time.After(time.Second):
			t.Error("write failure did not cancel connection context")
		}
	}))
	defer server.Close()

	dialURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), dialURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.CloseNow()
	<-accepted
	close(triggerWrite)

	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("connection context was not canceled")
	}
	select {
	case <-writeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("write loop did not exit")
	}
}
