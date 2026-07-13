package ws

import (
	"errors"
	"testing"
	"time"
)

func TestDebateStartLimiterEnforcesGlobalAndPerIP(t *testing.T) {
	limiter := NewDebateStartLimiter(DebateStartLimiterConfig{
		GlobalMax: 2,
		PerIPMax:  2,
		Window:    time.Minute,
	})
	now := time.Unix(1_700_000_000, 0)

	if err := limiter.Allow("1.1.1.1", now); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if err := limiter.Allow("2.2.2.2", now.Add(time.Second)); err != nil {
		t.Fatalf("second allow different IP: %v", err)
	}
	err := limiter.Allow("3.3.3.3", now.Add(2*time.Second))
	if err == nil {
		t.Fatal("expected global rate limit")
	}
	if !errors.Is(err, errDebateGlobalRateLimited) {
		t.Fatalf("error = %v", err)
	}
}

func TestDebateStartLimiterEnforcesPerIP(t *testing.T) {
	limiter := NewDebateStartLimiter(DebateStartLimiterConfig{
		GlobalMax: 10,
		PerIPMax:  1,
		Window:    time.Minute,
	})
	now := time.Unix(1_700_000_000, 0)

	if err := limiter.Allow("1.1.1.1", now); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	err := limiter.Allow("1.1.1.1", now.Add(time.Second))
	if err == nil {
		t.Fatal("expected per-IP rate limit")
	}
	if !errors.Is(err, errDebateIPRateLimited) {
		t.Fatalf("error = %v", err)
	}
	if err := limiter.Allow("2.2.2.2", now.Add(2*time.Second)); err != nil {
		t.Fatalf("other IP should be allowed: %v", err)
	}
}

func TestDebateStartLimiterWindowExpires(t *testing.T) {
	limiter := NewDebateStartLimiter(DebateStartLimiterConfig{
		GlobalMax: 1,
		PerIPMax:  1,
		Window:    time.Minute,
	})
	now := time.Unix(1_700_000_000, 0)

	if err := limiter.Allow("1.1.1.1", now); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if err := limiter.Allow("1.1.1.1", now.Add(61*time.Second)); err != nil {
		t.Fatalf("allow after window: %v", err)
	}
}

func TestClientIPFromRemoteAddr(t *testing.T) {
	if got := clientIP("192.0.2.10:54321"); got != "192.0.2.10" {
		t.Fatalf("clientIP = %q", got)
	}
	if got := clientIP("192.0.2.10"); got != "192.0.2.10" {
		t.Fatalf("clientIP without port = %q", got)
	}
}
