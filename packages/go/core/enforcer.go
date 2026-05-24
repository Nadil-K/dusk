package core

import (
	"fmt"
	"time"
)

type EnforcerResult struct {
	Enforce bool
	Body    map[string]string
}

func CheckSunset(ep EndpointConfig) EnforcerResult {
	if ep.SunsetAt == nil {
		return EnforcerResult{}
	}
	sunset, err := time.Parse("2006-01-02", *ep.SunsetAt)
	if err != nil || !time.Now().After(sunset) {
		return EnforcerResult{}
	}

	body := map[string]string{
		"error":   "Gone",
		"message": fmt.Sprintf("%s was sunset on %s and is no longer available.", ep.Path, *ep.SunsetAt),
	}
	if ep.Successor != nil {
		body["successor"] = *ep.Successor
	}
	if ep.MigrationDoc != nil {
		body["migration_doc"] = *ep.MigrationDoc
	}
	return EnforcerResult{Enforce: true, Body: body}
}
