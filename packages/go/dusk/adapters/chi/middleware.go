// Package chi provides dusk deprecation middleware for the chi router.
package chi

import (
	"encoding/json"
	"net/http"

	"github.com/nadilkarunarathna/dusk-go/dusk/core"
)

type DuskMiddleware struct {
	matcher *core.RouteMatcher
}

func New(endpoints []core.EndpointConfig) func(http.Handler) http.Handler {
	m := &DuskMiddleware{matcher: core.NewRouteMatcher(endpoints)}
	return m.Handler
}

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
		if result.Enforce {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			json.NewEncoder(w).Encode(result.Body)
			return
		}

		next.ServeHTTP(w, r)
	})
}
