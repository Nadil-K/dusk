package core

import "testing"

func makeTestEP(path string, methods ...string) EndpointConfig {
	if len(methods) == 0 {
		methods = []string{"GET"}
	}
	s := "2026-01-01"
	return EndpointConfig{Path: path, Methods: methods, DeprecatedAt: "2025-01-01", SunsetAt: &s}
}

func TestMatcher_ExactMatch(t *testing.T) {
	m := NewRouteMatcher([]EndpointConfig{makeTestEP("/api/v1/users")})
	if m.Match("/api/v1/users", "GET") == nil {
		t.Fatal("expected match")
	}
}

func TestMatcher_NoMatch_WrongPath(t *testing.T) {
	m := NewRouteMatcher([]EndpointConfig{makeTestEP("/api/v1/users")})
	if m.Match("/api/v2/users", "GET") != nil {
		t.Fatal("expected no match")
	}
}

func TestMatcher_NoMatch_WrongMethod(t *testing.T) {
	m := NewRouteMatcher([]EndpointConfig{makeTestEP("/api/v1/users", "POST")})
	if m.Match("/api/v1/users", "GET") != nil {
		t.Fatal("expected no match")
	}
}

func TestMatcher_ParamMatch(t *testing.T) {
	m := NewRouteMatcher([]EndpointConfig{makeTestEP("/api/v1/orders/{id}")})
	if m.Match("/api/v1/orders/123", "GET") == nil {
		t.Fatal("expected param match for numeric id")
	}
	if m.Match("/api/v1/orders/abc-xyz", "GET") == nil {
		t.Fatal("expected param match for string id")
	}
}

func TestMatcher_ParamNoMatch_ExtraSegments(t *testing.T) {
	m := NewRouteMatcher([]EndpointConfig{makeTestEP("/api/v1/orders/{id}")})
	if m.Match("/api/v1/orders/123/items", "GET") != nil {
		t.Fatal("expected no match for extra segments")
	}
}

func TestMatcher_MethodCaseInsensitive(t *testing.T) {
	m := NewRouteMatcher([]EndpointConfig{makeTestEP("/api/v1/users", "GET")})
	if m.Match("/api/v1/users", "get") == nil {
		t.Fatal("expected case-insensitive method match")
	}
}
