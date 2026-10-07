package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

var ServicePaths = map[string]bool{
	"/health":  true,
	"/ready":   true,
	"/metrics": true,
}

type ipLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimiter // ip -> limiter + lastSeen
	rps     rate.Limit
	burst   int
}
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var rateLimitedTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "gophprofile_http_rate_limited_total",
	Help: "Requests rejected with 429 by the rate limiter",
})

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		clients: make(map[string]*clientLimiter),
		rps:     rate.Limit(rps),
		burst:   burst,
	}
}

func WithRateLimit(rps float64, burst int) Middleware {
	l := newIPLimiter(rps, burst)
	go l.cleanup() // evict entries idle >3min, every minute — unbounded map otherwise
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ServicePaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			if !l.allow(clientIP(r)) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				rateLimitedTotal.Inc() // promauto counter
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	c, ok := l.clients[ip]
	if !ok {
		c = &clientLimiter{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.clients[ip] = c
	}
	c.lastSeen = time.Now()
	return c.limiter.Allow()
}

func (l *ipLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		l.mu.Lock()
		for ip, c := range l.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}
