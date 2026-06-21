package web

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter is a small per-key sliding-window limiter used to throttle the
// internet-facing cabinet auth endpoints (brute force + registration spam).
type rateLimiter struct {
	mu        sync.Mutex
	hits      map[string][]time.Time
	max       int
	window    time.Duration
	lastClean time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{hits: map[string][]time.Time{}, max: max, window: window, lastClean: time.Now()}
}

func (rl *rateLimiter) allow(key string) bool {
	now := time.Now()
	cut := now.Add(-rl.window)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if now.Sub(rl.lastClean) > rl.window {
		for k, ts := range rl.hits {
			if len(ts) == 0 || ts[len(ts)-1].Before(cut) {
				delete(rl.hits, k)
			}
		}
		rl.lastClean = now
	}
	ts := rl.hits[key]
	j := 0
	for _, t := range ts {
		if t.After(cut) {
			ts[j] = t
			j++
		}
	}
	ts = ts[:j]
	if len(ts) >= rl.max {
		rl.hits[key] = ts
		return false
	}
	rl.hits[key] = append(ts, now)
	return true
}

// clientIP returns the best-effort client IP, honoring a reverse proxy's
// forwarding headers (the bot is commonly behind nginx/Cloudflare).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// setSecurityHeaders applies baseline hardening headers. frameDeny is used for
// the cabinet (clickjacking protection); the Mini App is intentionally framable
// by Telegram, so it is not set there. HSTS is only meaningful over TLS.
func (s *Server) setSecurityHeaders(w http.ResponseWriter, frameDeny bool) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	if frameDeny {
		h.Set("X-Frame-Options", "DENY")
	}
	if s.domain != "" {
		h.Set("Strict-Transport-Security", "max-age=31536000")
	}
}
