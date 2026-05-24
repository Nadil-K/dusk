package store

import "time"

type HitEvent struct {
	Ts              time.Time
	Path            string
	Method          string
	CallerID        *string
	UserAgent       *string
	DaysUntilSunset *int
	EndpointKey     string
	Enforced        bool
}

type HitQuery struct {
	EndpointKey *string
	CallerID    *string
	SinceDays   int // 0 → default 30
	Limit       int // 0 → default 100
}

type TopCaller struct {
	CallerID string
	Count    int
}

type EndpointSummary struct {
	EndpointKey   string
	TotalHits     int
	UniqueCallers int
	LastSeen      *time.Time
	TopCallers    []TopCaller
}

type TotalSummary struct {
	TotalHits             int
	UniqueCallers         int
	EndpointsWithTraffic  int
	PastSunsetWithTraffic int
}

type HitStore interface {
	Record(hit HitEvent) error
	RecentHits(query HitQuery) ([]HitEvent, error)
	EndpointSummaries(sinceDays int) ([]EndpointSummary, error)
	TotalSummary(sinceDays int) (TotalSummary, error)
	Close() error
}
