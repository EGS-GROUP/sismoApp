package main

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Simple in-memory rate limiter (token bucket per IP)
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	tokens    int
	lastSeen  time.Time
	lastReset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok || now.Sub(v.lastReset) > rl.window {
		rl.visitors[ip] = &visitor{tokens: rl.limit - 1, lastSeen: now, lastReset: now}
		return true
	}

	if v.tokens <= 0 {
		v.lastSeen = now
		return false
	}

	v.tokens--
	v.lastSeen = now
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func rateLimitMiddleware(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded"}`))
			return
		}
		next(w, r)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "*"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func maxBodyMiddleware(maxBytes int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next(w, r)
	}
}

func allowedProxyDomain(targetURL string) bool {
	allowed := os.Getenv("PROXY_ALLOWED_DOMAINS")
	if allowed == "" {
		allowed = "rt-esp.rttv.com"
	}
	for _, domain := range strings.Split(allowed, ",") {
		domain = strings.TrimSpace(domain)
		if domain != "" && strings.Contains(targetURL, domain) {
			return true
		}
	}
	return false
}
