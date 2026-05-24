// Package chi provides dusk deprecation middleware for the chi router.
package chi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Nadil-K/dusk/packages/go/core"
	"github.com/Nadil-K/dusk/packages/go/core/store"
)

type DuskMiddleware struct {
	cfg     core.DuskConfig
	matcher *core.RouteMatcher
	store   *store.AsyncWriteBuffer
}

// New loads config from configPath and returns a chi-compatible middleware handler.
func New(configPath string) func(http.Handler) http.Handler {
	cfg, err := core.LoadConfig(configPath)
	if err != nil {
		panic("dusk: failed to load config: " + err.Error())
	}
	st, err := store.CreateStore(cfg)
	if err != nil {
		panic("dusk: failed to create store: " + err.Error())
	}
	m := &DuskMiddleware{
		cfg:     cfg,
		matcher: core.NewRouteMatcher(cfg.Endpoints),
		store:   st,
	}
	return m.Handler
}

// Handler returns a chi-compatible middleware function.
func (d *DuskMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ep := d.matcher.Match(r.URL.Path, r.Method)
		if ep == nil {
			next.ServeHTTP(w, r)
			return
		}

		headers := core.BuildHeaders(*ep)
		for k, v := range headers {
			w.Header().Set(k, v)
		}

		result := core.CheckSunset(*ep)
		days := core.DaysUntilSunset(*ep)
		callerID := d.resolveCallerID(r)

		ua := r.Header.Get("User-Agent")
		var userAgent *string
		if ua != "" {
			userAgent = &ua
		}

		hit := store.HitEvent{
			Ts:              time.Now().UTC(),
			Path:            r.URL.Path,
			Method:          r.Method,
			CallerID:        callerID,
			UserAgent:       userAgent,
			DaysUntilSunset: days,
			EndpointKey:     ep.Path,
			Enforced:        result.Enforce,
		}
		if err := d.store.Record(hit); err != nil {
			slog.Warn("dusk: failed to record hit", "error", err)
		}

		if result.Enforce {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			json.NewEncoder(w).Encode(result.Body) //nolint:errcheck
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (d *DuskMiddleware) resolveCallerID(r *http.Request) *string {
	for _, rule := range d.cfg.Log.IdentifyBy {
		switch v := rule.(type) {
		case map[string]interface{}:
			if header, ok := v["header"].(string); ok {
				if val := r.Header.Get(header); val != "" {
					return &val
				}
			}
		case string:
			if v == "ip" {
				ip := r.RemoteAddr
				if i := strings.LastIndex(ip, ":"); i != -1 {
					ip = ip[:i]
				}
				if ip != "" {
					return &ip
				}
			}
		}
	}
	return nil
}
