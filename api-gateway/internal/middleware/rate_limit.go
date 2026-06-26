package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientState struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientState
	rate    float64 // tokens per second
	burst   float64 // max tokens
}

func NewRateLimiter(r float64, b float64) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientState),
		rate:    r,
		burst:   b,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, client := range rl.clients {
			if now.Sub(client.lastSeen) > 5*time.Minute {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getClientKey(r *http.Request) string {
	// 1. Identify by X-User-ID (passed by Auth middleware)
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return "user:" + userID
	}
	// 2. Fallback to IP address (e.g. from X-Forwarded-For or RemoteAddr)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return "ip:" + strings.TrimSpace(parts[0])
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "ip:" + r.RemoteAddr
	}
	return "ip:" + ip
}

func (rl *RateLimiter) Allow(r *http.Request) bool {
	key := rl.getClientKey(r)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[key]
	if !exists {
		client = &clientState{
			tokens:   rl.burst,
			lastSeen: now,
		}
		rl.clients[key] = client
	}

	elapsed := now.Sub(client.lastSeen).Seconds()
	client.lastSeen = now
	client.tokens += elapsed * rl.rate
	if client.tokens > rl.burst {
		client.tokens = rl.burst
	}

	if client.tokens >= 1.0 {
		client.tokens -= 1.0
		return true
	}
	return false
}

func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"too many requests"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
