package middleware

import (
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultLoginMaxFailures = 5
	defaultLoginWindow      = 3 * time.Minute

	loginSweepInterval = 3 * time.Minute

	// Number of proxies between the client and this process whose addresses
	// must be skipped when reading X-Forwarded-For. Caddy appends the address
	// of its own peer, so the last entry is never the client:
	//
	//   client -> platform LB -> Caddy -> here
	//   X-Forwarded-For: <client>, <platform LB>
	//
	// Skipping one entry from the right therefore yields the real client, and a
	// header forged by the client is pushed further left where it is ignored.
	// Override with LOGIN_RATELIMIT_TRUSTED_HOPS if the topology differs.
	defaultTrustedProxyHops = 1
)

type loginFailures struct {
	count      int
	windowEnds time.Time
}

type LoginLimiter struct {
	mu        sync.Mutex
	failures  map[string]*loginFailures
	lastSweep time.Time

	maxFailures int
	window      time.Duration
	trustedHops int

	now func() time.Time
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{
		failures:    make(map[string]*loginFailures),
		maxFailures: defaultLoginMaxFailures,
		window:      defaultLoginWindow,
		trustedHops: trustedProxyHopsFromEnv(),
		now:         time.Now,
	}
}

func trustedProxyHopsFromEnv() int {
	raw := os.Getenv("LOGIN_RATELIMIT_TRUSTED_HOPS")
	if raw == "" {
		return defaultTrustedProxyHops
	}

	hops, err := strconv.Atoi(raw)
	if err != nil || hops < 0 {
		log.Printf("Invalid LOGIN_RATELIMIT_TRUSTED_HOPS %q, falling back to %d", raw, defaultTrustedProxyHops)
		return defaultTrustedProxyHops
	}
	return hops
}

func (ll *LoginLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := ll.clientKey(r)

		if retryAfter, locked := ll.locked(client); locked {
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
			http.Error(w, "Too many failed login attempts, please try again later", http.StatusTooManyRequests)
			return
		}

		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		switch rec.Status() {
		case http.StatusUnauthorized:
			ll.recordFailure(client)
		case http.StatusOK:
			ll.reset(client)
		}
	})
}

func (ll *LoginLimiter) locked(client string) (time.Duration, bool) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	now := ll.now()
	ll.sweep(now)

	f, ok := ll.failures[client]
	if !ok || !now.Before(f.windowEnds) || f.count < ll.maxFailures {
		return 0, false
	}
	return f.windowEnds.Sub(now), true
}

func (ll *LoginLimiter) recordFailure(client string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	now := ll.now()
	f, ok := ll.failures[client]
	if !ok || !now.Before(f.windowEnds) {
		ll.failures[client] = &loginFailures{count: 1, windowEnds: now.Add(ll.window)}
		return
	}

	f.count++
	f.windowEnds = now.Add(ll.window)
}

func (ll *LoginLimiter) reset(client string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	delete(ll.failures, client)
}

func (ll *LoginLimiter) sweep(now time.Time) {
	if now.Sub(ll.lastSweep) < loginSweepInterval {
		return
	}
	ll.lastSweep = now

	for client, f := range ll.failures {
		if !now.Before(f.windowEnds) {
			delete(ll.failures, client)
		}
	}
}

func (ll *LoginLimiter) clientKey(r *http.Request) string {
	forwarded := r.Header.Values("X-Forwarded-For")
	var hops []string
	for _, header := range forwarded {
		for _, entry := range strings.Split(header, ",") {
			if entry = strings.TrimSpace(entry); entry != "" {
				hops = append(hops, entry)
			}
		}
	}

	if candidate := len(hops) - 1 - ll.trustedHops; candidate >= 0 {
		if ip := net.ParseIP(strings.Trim(hops[candidate], "[]")); ip != nil {
			return ip.String()
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
