package core

import "testing"

func TestCheckSunset_NoSunset(t *testing.T) {
	ep := EndpointConfig{Path: "/api/v1/users", DeprecatedAt: "2025-01-01"}
	if CheckSunset(ep).Enforce {
		t.Fatal("expected no enforcement when no sunset configured")
	}
}

func TestCheckSunset_FutureSunset(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		DeprecatedAt: "2025-01-01",
		SunsetAt:     strPtr("2099-01-01"),
	}
	if CheckSunset(ep).Enforce {
		t.Fatal("expected no enforcement for future sunset")
	}
}

func TestCheckSunset_PastSunset(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		DeprecatedAt: "2025-01-01",
		SunsetAt:     strPtr("2020-01-01"),
	}
	r := CheckSunset(ep)
	if !r.Enforce {
		t.Fatal("expected enforcement for past sunset")
	}
	if r.Body["error"] == "" {
		t.Fatal("expected non-empty error body")
	}
}

func TestCheckSunset_PastSunset_SuccessorInBody(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		DeprecatedAt: "2025-01-01",
		SunsetAt:     strPtr("2020-01-01"),
		Successor:    strPtr("/api/v2/users"),
	}
	r := CheckSunset(ep)
	if !r.Enforce {
		t.Fatal("expected enforcement")
	}
	if r.Body["successor"] != "/api/v2/users" {
		t.Fatalf("expected successor in body, got %v", r.Body)
	}
}
