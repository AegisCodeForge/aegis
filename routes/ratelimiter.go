package routes

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/bctnry/aegis/pkg/aegis"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter map[string]*rate.Limiter
	mutex *sync.RWMutex
	limit rate.Limit
	cap int
}

func NewRateLimiter(cfg *aegis.AegisConfig) *RateLimiter {
	return &RateLimiter{
		limiter: make(map[string]*rate.Limiter, 0),
		mutex: &sync.RWMutex{},
		limit: rate.Limit(cfg.MaxRequestInSecond),
		cap: 1,
	}
}

func ResolveMostPossibleIP(w http.ResponseWriter, r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	netip := net.ParseIP(ip)
	if netip != nil { return netip.String() }

	ips := r.Header.Values("X-Forwarded-For")
	// one should know that this is never going to be 100% correct.
	for _, ip := range ips {
		for k := range strings.SplitSeq(ip, ",") {
			netip = net.ParseIP(strings.TrimSpace(k))
			if netip != nil { return netip.String() }
		}
	}

	// TODO: handle "Forward" header as well.

	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil { return r.RemoteAddr }
	return h
}

func (rl *RateLimiter) IsIPAllowed(s string) bool {
	var r *rate.Limiter
	var ok bool
	r, ok = rl.limiter[s]
	if !ok {
		rl.mutex.Lock()
		rl.limiter[s] = rate.NewLimiter(rl.limit, rl.cap)
		r = rl.limiter[s]
		rl.mutex.Unlock()
	}
	return r.Allow()
}

