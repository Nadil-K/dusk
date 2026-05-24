package core

import (
	"fmt"
	"strings"
	"time"
)

type DeprecationHeaders map[string]string

func BuildHeaders(ep EndpointConfig) DeprecationHeaders {
	h := DeprecationHeaders{
		"Deprecation": fmt.Sprintf(`@"%s"`, ep.DeprecatedAt),
	}
	if ep.SunsetAt != nil {
		h["Sunset"] = *ep.SunsetAt
	}
	var links []string
	if ep.Successor != nil {
		links = append(links, fmt.Sprintf(`<%s>; rel="successor-version"`, *ep.Successor))
	}
	if ep.MigrationDoc != nil {
		links = append(links, fmt.Sprintf(`<%s>; rel="deprecation"`, *ep.MigrationDoc))
	}
	if len(links) > 0 {
		h["Link"] = strings.Join(links, ", ")
	}
	return h
}

func DaysUntilSunset(ep EndpointConfig) *int {
	if ep.SunsetAt == nil {
		return nil
	}
	sunset, err := time.Parse("2006-01-02", *ep.SunsetAt)
	if err != nil {
		return nil
	}
	days := int(time.Until(sunset).Hours() / 24)
	return &days
}
