package core

import (
	"regexp"
	"strings"
)

var paramRE = regexp.MustCompile(`\{[^}]+\}`)

func pathToRegex(path string) *regexp.Regexp {
	parts := paramRE.Split(path, -1)
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)
	}
	return regexp.MustCompile(`^` + strings.Join(parts, `[^/]+`) + `$`)
}

type RouteMatcher struct {
	routes []struct {
		pattern  *regexp.Regexp
		endpoint EndpointConfig
	}
}

func NewRouteMatcher(endpoints []EndpointConfig) *RouteMatcher {
	rm := &RouteMatcher{}
	for _, ep := range endpoints {
		rm.routes = append(rm.routes, struct {
			pattern  *regexp.Regexp
			endpoint EndpointConfig
		}{pathToRegex(ep.Path), ep})
	}
	return rm
}

func (rm *RouteMatcher) Match(path, method string) *EndpointConfig {
	upper := strings.ToUpper(method)
	for _, r := range rm.routes {
		if r.pattern.MatchString(path) {
			for _, m := range r.endpoint.Methods {
				if strings.ToUpper(m) == upper {
					ep := r.endpoint
					return &ep
				}
			}
		}
	}
	return nil
}
