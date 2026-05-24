package core

import (
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestBuildHeaders_Deprecation(t *testing.T) {
	ep := EndpointConfig{Path: "/api/v1/users", Methods: []string{"GET"}, DeprecatedAt: "2025-01-01"}
	h := BuildHeaders(ep)
	if h["Deprecation"] != "@1735689600" {
		t.Fatalf("expected Deprecation=@1735689600, got %q", h["Deprecation"])
	}
}

func TestBuildHeaders_Sunset(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		Methods:      []string{"GET"},
		DeprecatedAt: "2025-01-01",
		SunsetAt:     strPtr("2026-06-01"),
	}
	h := BuildHeaders(ep)
	if h["Sunset"] != "Mon, 01 Jun 2026 00:00:00 GMT" {
		t.Fatalf("expected Sunset=Mon, 01 Jun 2026 00:00:00 GMT, got %q", h["Sunset"])
	}
}

func TestBuildHeaders_LinkSuccessor(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		Methods:      []string{"GET"},
		DeprecatedAt: "2025-01-01",
		Successor:    strPtr("/api/v2/users"),
	}
	h := BuildHeaders(ep)
	if !strings.Contains(h["Link"], "successor-version") {
		t.Fatalf("expected Link with successor-version, got %q", h["Link"])
	}
}

func TestBuildHeaders_LinkMigrationDoc(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		Methods:      []string{"GET"},
		DeprecatedAt: "2025-01-01",
		MigrationDoc: strPtr("https://docs.example.com/migration"),
	}
	h := BuildHeaders(ep)
	if !strings.Contains(h["Link"], "deprecation") {
		t.Fatalf("expected Link with deprecation rel, got %q", h["Link"])
	}
}

func TestDaysUntilSunset_NoSunset(t *testing.T) {
	ep := EndpointConfig{Path: "/api/v1/users", DeprecatedAt: "2025-01-01"}
	if DaysUntilSunset(ep) != nil {
		t.Fatal("expected nil when no sunset")
	}
}

func TestDaysUntilSunset_FutureSunset(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		DeprecatedAt: "2025-01-01",
		SunsetAt:     strPtr("2099-01-01"),
	}
	days := DaysUntilSunset(ep)
	if days == nil || *days <= 0 {
		t.Fatal("expected positive days for future sunset")
	}
}

func TestDaysUntilSunset_PastSunset(t *testing.T) {
	ep := EndpointConfig{
		Path:         "/api/v1/users",
		DeprecatedAt: "2025-01-01",
		SunsetAt:     strPtr("2020-01-01"),
	}
	days := DaysUntilSunset(ep)
	if days == nil || *days >= 0 {
		t.Fatal("expected negative days for past sunset")
	}
}
