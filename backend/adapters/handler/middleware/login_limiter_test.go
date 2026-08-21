package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLimiter(maxFailures int, window time.Duration) (*LoginLimiter, *time.Time) {
	clock := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	ll := &LoginLimiter{
		failures:    make(map[string]*loginFailures),
		maxFailures: maxFailures,
		window:      window,
		trustedHops: defaultTrustedProxyHops,
		now:         func() time.Time { return clock },
	}
	return ll, &clock
}

func loginStub(statuses ...int) http.Handler {
	call := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := statuses[len(statuses)-1]
		if call < len(statuses) {
			status = statuses[call]
		}
		call++
		w.WriteHeader(status)
	})
}

func attempt(handler http.Handler, remoteAddr string, forwardedFor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/auth/login", nil)
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestLoginLimiter(t *testing.T) {
	t.Run("failed attempts below the threshold are passed through", func(t *testing.T) {
		ll, _ := newTestLimiter(3, 15*time.Minute)
		handler := ll.Handle(loginStub(http.StatusUnauthorized))

		for i := 0; i < 2; i++ {
			if got := attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5").Code; got != http.StatusUnauthorized {
				t.Fatalf("attempt %d: expected 401, got %d", i+1, got)
			}
		}
	})

	t.Run("the attempt after the threshold is rejected with 429 and Retry-After", func(t *testing.T) {
		ll, _ := newTestLimiter(3, 15*time.Minute)
		handler := ll.Handle(loginStub(http.StatusUnauthorized))

		for i := 0; i < 3; i++ {
			attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5")
		}

		rr := attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5")
		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", rr.Code)
		}
		if got := rr.Header().Get("Retry-After"); got != "900" {
			t.Errorf("expected Retry-After 900, got %q", got)
		}
	})

	t.Run("a locked out client is let in again once the window has passed", func(t *testing.T) {
		ll, clock := newTestLimiter(3, 15*time.Minute)
		handler := ll.Handle(loginStub(http.StatusUnauthorized))

		for i := 0; i < 3; i++ {
			attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5")
		}
		if got := attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5").Code; got != http.StatusTooManyRequests {
			t.Fatalf("expected the client to be locked out, got %d", got)
		}

		*clock = clock.Add(15*time.Minute + time.Second)

		if got := attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5").Code; got != http.StatusUnauthorized {
			t.Errorf("expected the handler to be reached again, got %d", got)
		}
	})

	t.Run("a successful login clears the failure count", func(t *testing.T) {
		ll, _ := newTestLimiter(3, 15*time.Minute)
		handler := ll.Handle(loginStub(http.StatusUnauthorized, http.StatusUnauthorized, http.StatusOK, http.StatusUnauthorized))

		attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5") // 401
		attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5") // 401
		attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5") // 200 -> reset

		if got := attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5").Code; got != http.StatusUnauthorized {
			t.Errorf("expected the budget to be reset, got %d", got)
		}
	})

	t.Run("clients are counted separately", func(t *testing.T) {
		ll, _ := newTestLimiter(3, 15*time.Minute)
		handler := ll.Handle(loginStub(http.StatusUnauthorized))

		for i := 0; i < 4; i++ {
			attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5")
		}

		if got := attempt(handler, "10.0.0.5:1234", "9.8.7.6, 10.0.0.5").Code; got != http.StatusUnauthorized {
			t.Errorf("expected a different client to be unaffected, got %d", got)
		}
	})

	t.Run("expired counters are swept", func(t *testing.T) {
		ll, clock := newTestLimiter(3, 15*time.Minute)
		handler := ll.Handle(loginStub(http.StatusUnauthorized))

		attempt(handler, "10.0.0.5:1234", "1.2.3.4, 10.0.0.5")
		if len(ll.failures) != 1 {
			t.Fatalf("expected 1 tracked client, got %d", len(ll.failures))
		}

		*clock = clock.Add(15*time.Minute + loginSweepInterval)
		attempt(handler, "10.0.0.5:1234", "9.8.7.6, 10.0.0.5")

		if _, tracked := ll.failures["1.2.3.4"]; tracked {
			t.Error("expected the expired counter to be dropped")
		}
	})
}

func TestLoginLimiterClientKey(t *testing.T) {
	ll, _ := newTestLimiter(3, 15*time.Minute)

	t.Run("the address appended by Caddy is skipped", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		req.RemoteAddr = "127.0.0.1:5000"
		req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.5")

		if got := ll.clientKey(req); got != "1.2.3.4" {
			t.Errorf("expected 1.2.3.4, got %q", got)
		}
	})

	t.Run("a header forged by the client is ignored", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		req.RemoteAddr = "127.0.0.1:5000"
		req.Header.Set("X-Forwarded-For", "9.9.9.9, 1.2.3.4, 10.0.0.5")

		if got := ll.clientKey(req); got != "1.2.3.4" {
			t.Errorf("expected 1.2.3.4, got %q", got)
		}
	})

	t.Run("without a forwarded header the peer address is used", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		req.RemoteAddr = "192.168.1.20:5000"

		if got := ll.clientKey(req); got != "192.168.1.20" {
			t.Errorf("expected 192.168.1.20, got %q", got)
		}
	})
}
