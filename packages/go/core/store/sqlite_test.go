package store

import (
	"path/filepath"
	"testing"
	"time"
)

func makeHit(overrides ...func(*HitEvent)) HitEvent {
	callerID := "key-abc"
	ua := "test/1.0"
	days := 100
	h := HitEvent{
		Ts:              time.Now().UTC(),
		Path:            "/api/v1/users",
		Method:          "GET",
		CallerID:        &callerID,
		UserAgent:       &ua,
		DaysUntilSunset: &days,
		EndpointKey:     "/api/v1/users",
		Enforced:        false,
	}
	for _, o := range overrides {
		o(&h)
	}
	return h
}

func newTestSQLite(t *testing.T) *SQLiteHitStore {
	t.Helper()
	s, err := NewSQLiteHitStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSQLite_RecordAndRecentHits(t *testing.T) {
	s := newTestSQLite(t)
	if err := s.Record(makeHit()); err != nil {
		t.Fatal(err)
	}
	hits, err := s.RecentHits(HitQuery{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Path != "/api/v1/users" {
		t.Fatalf("unexpected path: %s", hits[0].Path)
	}
}

func TestSQLite_EnforcedPersisted(t *testing.T) {
	s := newTestSQLite(t)
	s.Record(makeHit(func(h *HitEvent) { h.Enforced = true }))
	hits, _ := s.RecentHits(HitQuery{})
	if !hits[0].Enforced {
		t.Fatal("expected enforced=true")
	}
}

func TestSQLite_FilterByEndpointKey(t *testing.T) {
	s := newTestSQLite(t)
	ep2 := "/api/v1/orders/{id}"
	s.Record(makeHit())
	s.Record(makeHit(func(h *HitEvent) { h.Path = "/api/v1/orders/1"; h.EndpointKey = ep2 }))

	ep := "/api/v1/users"
	hits, _ := s.RecentHits(HitQuery{EndpointKey: &ep})
	if len(hits) != 1 || hits[0].EndpointKey != ep {
		t.Fatalf("expected 1 hit for /api/v1/users, got %d", len(hits))
	}
}

func TestSQLite_EndpointSummaries(t *testing.T) {
	s := newTestSQLite(t)
	callerA := "key-a"
	callerB := "key-b"
	ep2 := "/api/v1/orders/{id}"
	s.Record(makeHit(func(h *HitEvent) { h.CallerID = &callerA }))
	s.Record(makeHit(func(h *HitEvent) { h.CallerID = &callerB }))
	s.Record(makeHit(func(h *HitEvent) {
		h.Path = "/api/v1/orders/1"
		h.EndpointKey = ep2
		h.CallerID = &callerA
	}))

	summaries, err := s.EndpointSummaries(30)
	if err != nil {
		t.Fatal(err)
	}
	var users *EndpointSummary
	for i := range summaries {
		if summaries[i].EndpointKey == "/api/v1/users" {
			users = &summaries[i]
		}
	}
	if users == nil {
		t.Fatal("expected /api/v1/users in summaries")
	}
	if users.TotalHits != 2 {
		t.Fatalf("expected TotalHits=2, got %d", users.TotalHits)
	}
	if users.UniqueCallers != 2 {
		t.Fatalf("expected UniqueCallers=2, got %d", users.UniqueCallers)
	}
}

func TestSQLite_TotalSummary(t *testing.T) {
	s := newTestSQLite(t)
	callerA := "key-a"
	daysPos := 10
	callerB := "key-b"
	daysPast := -5
	ep2 := "/api/v1/gone"

	s.Record(makeHit(func(h *HitEvent) { h.CallerID = &callerA; h.DaysUntilSunset = &daysPos }))
	s.Record(makeHit(func(h *HitEvent) {
		h.Path = ep2; h.EndpointKey = ep2
		h.CallerID = &callerB; h.DaysUntilSunset = &daysPast; h.Enforced = true
	}))

	total, err := s.TotalSummary(30)
	if err != nil {
		t.Fatal(err)
	}
	if total.TotalHits != 2 {
		t.Fatalf("expected TotalHits=2, got %d", total.TotalHits)
	}
	if total.EndpointsWithTraffic != 2 {
		t.Fatalf("expected EndpointsWithTraffic=2, got %d", total.EndpointsWithTraffic)
	}
	if total.PastSunsetWithTraffic != 1 {
		t.Fatalf("expected PastSunsetWithTraffic=1, got %d", total.PastSunsetWithTraffic)
	}
}

func TestSQLite_OrdersNewestFirst(t *testing.T) {
	s := newTestSQLite(t)
	now := time.Now().UTC()
	s.Record(makeHit(func(h *HitEvent) { h.Ts = now.Add(-10 * time.Second) }))
	s.Record(makeHit(func(h *HitEvent) { h.Ts = now.Add(-5 * time.Second) }))

	hits, _ := s.RecentHits(HitQuery{Limit: 10})
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
	if hits[0].Ts.Before(hits[1].Ts) {
		t.Fatal("expected newest hit first")
	}
}
