package chi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nadil-K/dusk/packages/go/core"
	"github.com/Nadil-K/dusk/packages/go/core/store"
)

const testConfigTmpl = `
dusk:
  version: 1
  store:
    backend: sqlite
    path: %s
  log:
    identify_by:
      - header: X-API-Key
  endpoints:
    - path: /api/v1/users
      methods: [GET]
      deprecated_at: "2025-01-01"
      sunset_at: "2099-01-01"
      successor: /api/v2/users
    - path: /api/v1/gone
      methods: [GET]
      deprecated_at: "2024-01-01"
      sunset_at: "2020-01-01"
`

func newTestMiddleware(t *testing.T) (*DuskMiddleware, func(http.Handler) http.Handler) {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "dusk.yaml")
	dbPath := filepath.Join(tmpDir, "test.db")
	if err := os.WriteFile(cfgPath, []byte(fmt.Sprintf(testConfigTmpl, dbPath)), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.CreateStore(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := &DuskMiddleware{
		cfg:     cfg,
		matcher: core.NewRouteMatcher(cfg.Endpoints),
		store:   st,
	}
	t.Cleanup(func() { m.store.Close() })
	return m, m.Handler
}

func serve(handler func(http.Handler) http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	return rr
}

func TestMiddleware_NoMatch_PassThrough(t *testing.T) {
	_, handler := newTestMiddleware(t)
	rr := serve(handler, "GET", "/api/v2/users", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Deprecation") != "" {
		t.Fatal("expected no Deprecation header for non-deprecated endpoint")
	}
}

func TestMiddleware_Deprecated_HeadersSet(t *testing.T) {
	_, handler := newTestMiddleware(t)
	rr := serve(handler, "GET", "/api/v1/users", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Deprecation"); got != "@1735689600" {
		t.Fatalf("expected Deprecation=@1735689600, got %q", got)
	}
	if got := rr.Header().Get("Sunset"); got != "Thu, 01 Jan 2099 00:00:00 GMT" {
		t.Fatalf("expected Sunset=Thu, 01 Jan 2099 00:00:00 GMT, got %q", got)
	}
}

func TestMiddleware_Sunset_Returns410(t *testing.T) {
	_, handler := newTestMiddleware(t)
	rr := serve(handler, "GET", "/api/v1/gone", nil)
	if rr.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatal("expected JSON content-type on 410")
	}
}

func TestMiddleware_HitRecorded(t *testing.T) {
	m, handler := newTestMiddleware(t)
	serve(handler, "GET", "/api/v1/users", map[string]string{"X-API-Key": "test-key"})

	m.store.FlushAll()

	ep := "/api/v1/users"
	hits, err := m.store.RecentHits(store.HitQuery{EndpointKey: &ep})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected at least one hit recorded")
	}
	if hits[0].CallerID == nil || *hits[0].CallerID != "test-key" {
		t.Fatalf("expected caller_id=test-key, got %v", hits[0].CallerID)
	}
}

func TestMiddleware_SunsetHitLogged(t *testing.T) {
	m, handler := newTestMiddleware(t)
	rr := serve(handler, "GET", "/api/v1/gone", nil)
	if rr.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", rr.Code)
	}

	m.store.FlushAll()

	ep := "/api/v1/gone"
	hits, err := m.store.RecentHits(store.HitQuery{EndpointKey: &ep})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected sunset hit to be logged even on 410")
	}
	if !hits[0].Enforced {
		t.Fatal("expected hit.Enforced=true for 410 response")
	}
}
