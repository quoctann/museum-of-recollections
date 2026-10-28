package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"museum-of-recollections/apps/api/internal/health"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type readyFunc func(context.Context) error

type dependencyCheck func(context.Context) error

func (f dependencyCheck) Check(ctx context.Context) error { return f(ctx) }

func TestReadinessDiagnosticsAreRedacted(t *testing.T) {
	for _, tc := range []struct {
		err    error
		reason string
	}{
		{errors.New("postgres://private:password@host/database"), "unavailable"},
		{context.DeadlineExceeded, "timeout"},
		{context.Canceled, "canceled"},
	} {
		core, logs := observer.New(zap.WarnLevel)
		service := health.New(time.Second, dependencyCheck(func(context.Context) error { return tc.err }))
		h := New(service, zap.New(core))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), "private") {
			t.Fatalf("unsafe response: %s", w.Body.String())
		}
		entries := logs.All()
		if len(entries) != 1 {
			t.Fatalf("expected one diagnostic, got %d", len(entries))
		}
		fields := entries[0].ContextMap()
		if len(fields) != 2 || fields["dependency_index"] != int64(0) || fields["failure_reason"] != tc.reason {
			t.Fatalf("unexpected diagnostic fields: %v", fields)
		}
		if entries[0].Message != "readiness check failed" {
			t.Fatal("unexpected log message")
		}
	}
}

func TestRecoverySanitizesPanic(t *testing.T) {
	r := gin.New()
	r.Use(recovery(zap.NewNop()))
	r.GET("/panic", func(c *gin.Context) { panic("private secret") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "private") {
		t.Fatalf("unsafe panic response: %d %s", w.Code, w.Body.String())
	}
}

func (f readyFunc) Ready(c context.Context) error { return f(c) }
func TestRoutes(t *testing.T) {
	for _, fail := range []bool{false, true} {
		h := New(readyFunc(func(ctx context.Context) error {
			if _, ok := ctx.Deadline(); !ok {
				t.Error("missing request deadline")
			}
			if fail {
				return errors.New("password private")
			}
			return nil
		}), zap.NewNop())
		for _, path := range []string{"/healthz", "/readyz", "/missing"} {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			want := 200
			if path == "/missing" {
				want = 404
			}
			if path == "/readyz" && fail {
				want = 503
			}
			if w.Code != want {
				t.Fatalf("%s got %d", path, w.Code)
			}
			if strings.Contains(w.Body.String(), "private") {
				t.Fatal("details leaked")
			}
		}
	}
}
func TestRateLimitAllRoutes(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz", "/missing"} {
		h := New(readyFunc(func(context.Context) error { return nil }), zap.NewNop())
		limited := false
		for i := 0; i < 100; i++ {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code == 429 {
				limited = true
				if w.Header().Get("Retry-After") == "" {
					t.Error("missing retry header")
				}
			}
		}
		if !limited {
			t.Fatal("not limited")
		}
	}
}
func TestProbeBudgetsAreIndependent(t *testing.T) {
	for _, attack := range []struct{ method, path string }{
		{http.MethodGet, "/missing"},
		{http.MethodPost, "/healthz"},
		{http.MethodGet, "/readyz"},
		{http.MethodGet, "/healthz"},
	} {
		t.Run(attack.method+attack.path, func(t *testing.T) {
			h := New(readyFunc(func(context.Context) error { return nil }), zap.NewNop())
			for i := 0; i < 100; i++ {
				h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(attack.method, attack.path, nil))
			}
			for _, probe := range []string{"/healthz", "/readyz"} {
				if attack.method == http.MethodGet && attack.path == probe {
					continue
				}
				w := httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, probe, nil))
				if w.Code != http.StatusOK {
					t.Fatalf("%s exhausted %s: status %d", attack.path, probe, w.Code)
				}
			}
		})
	}
}
func TestLimiterRefillsAndConcurrent(t *testing.T) {
	l := newLimiter(1, 1)
	now := time.Now()
	if !l.allow(now) || l.allow(now) || !l.allow(now.Add(time.Second)) {
		t.Fatal("token behavior")
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); l.allow(time.Now()) }()
	}
	wg.Wait()
}
func TestCanceledRequest(t *testing.T) {
	h := New(readyFunc(func(c context.Context) error { return c.Err() }), zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil).WithContext(ctx))
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}

func TestArticleRoutesRemainUnavailableBeforeAuthentication(t *testing.T) {
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/public/articles"},
		{http.MethodPost, "/api/admin/articles"},
		{http.MethodGet, "/api/admin/articles/example"},
		{http.MethodPut, "/api/admin/articles/example/draft"},
		{http.MethodPost, "/api/admin/articles/example/publish"},
		{http.MethodPost, "/api/admin/articles/example/unpublish"},
		{http.MethodPost, "/api/admin/media"},
	} {
		t.Run(route.method+route.path, func(t *testing.T) {
			h := New(readyFunc(func(context.Context) error { return nil }), zap.NewNop())
			response := httptest.NewRecorder()
			h.ServeHTTP(response, httptest.NewRequest(route.method, route.path, nil))
			if response.Code != http.StatusNotFound {
				t.Fatalf("route must remain unregistered, got %d", response.Code)
			}
		})
	}
}
