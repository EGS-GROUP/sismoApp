package main

import (
	"os"
	"testing"
	"time"
)

func TestAllowedProxyDomain(t *testing.T) {
	os.Setenv("PROXY_ALLOWED_DOMAINS", "rt-esp.rttv.com,example.com")
	defer os.Unsetenv("PROXY_ALLOWED_DOMAINS")

	if !allowedProxyDomain("https://rt-esp.rttv.com/live/index.m3u8") {
		t.Fatal("expected rt-esp.rttv.com to be allowed")
	}
	if !allowedProxyDomain("https://example.com/stream.m3u8") {
		t.Fatal("expected example.com to be allowed")
	}
	if allowedProxyDomain("https://evil.com/stream.m3u8") {
		t.Fatal("expected evil.com to be denied")
	}
}

func TestRateLimiter(t *testing.T) {
	// In CI the cleanup goroutine may run; use a small window to keep test fast
	// but avoid flakiness from cleanup between requests by using a large window
	// and not relying on it for this test.
	const ip = "127.0.0.1"
	// Limit of 3 requests per hour
	rl := newRateLimiter(3, time.Hour)

	if !rl.allow(ip) {
		t.Fatal("first request should be allowed")
	}
	if !rl.allow(ip) {
		t.Fatal("second request should be allowed")
	}
	if !rl.allow(ip) {
		t.Fatal("third request should be allowed")
	}
	if rl.allow(ip) {
		t.Fatal("fourth request should be denied")
	}
}
