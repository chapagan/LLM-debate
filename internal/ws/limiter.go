package ws

import (
	"errors"
	"net"
	"sync"
	"time"
)

type DebateStartLimiterConfig struct {
	GlobalMax int
	PerIPMax  int
	Window    time.Duration
}

type DebateStartLimiter struct {
	mu        sync.Mutex
	globalMax int
	perIPMax  int
	window    time.Duration
	global    []time.Time
	perIP     map[string][]time.Time
}

const (
	defaultGlobalDebateStartsPerWindow = 6
	defaultPerIPDebateStartsPerWindow  = 2
	defaultDebateStartWindow           = time.Minute
)

var (
	errDebateGlobalRateLimited = errors.New("debate start globally rate limited; wait before starting another debate")
	errDebateIPRateLimited     = errors.New("debate start rate limited for this address; wait before starting another debate")
)

func NewDebateStartLimiter(cfg DebateStartLimiterConfig) *DebateStartLimiter {
	globalMax := cfg.GlobalMax
	if globalMax <= 0 {
		globalMax = defaultGlobalDebateStartsPerWindow
	}
	perIPMax := cfg.PerIPMax
	if perIPMax <= 0 {
		perIPMax = defaultPerIPDebateStartsPerWindow
	}
	window := cfg.Window
	if window <= 0 {
		window = defaultDebateStartWindow
	}
	return &DebateStartLimiter{
		globalMax: globalMax,
		perIPMax:  perIPMax,
		window:    window,
		perIP:     make(map[string][]time.Time),
	}
}

func (l *DebateStartLimiter) Allow(ip string, now time.Time) error {
	if l == nil {
		return nil
	}
	if ip == "" {
		ip = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	l.global = pruneTimes(l.global, cutoff)
	l.perIP[ip] = pruneTimes(l.perIP[ip], cutoff)

	if len(l.global) >= l.globalMax {
		return errDebateGlobalRateLimited
	}
	if len(l.perIP[ip]) >= l.perIPMax {
		return errDebateIPRateLimited
	}

	l.global = append(l.global, now)
	l.perIP[ip] = append(l.perIP[ip], now)
	return nil
}

func pruneTimes(times []time.Time, cutoff time.Time) []time.Time {
	kept := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	return kept
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
