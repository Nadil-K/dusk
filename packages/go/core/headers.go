package core

import (
	"fmt"
	"strings"
	"time"
)

type DeprecationHeaders map[string]string

func dateToUnix(isoDate string) int64 {
	t, _ := time.Parse("2006-01-02", isoDate)
	return t.UTC().Unix()
}

func dateToHTTPDate(isoDate string) string {
	t, _ := time.Parse("2006-01-02", isoDate)
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"
}

func BuildHeaders(ep EndpointConfig) DeprecationHeaders {
	h := DeprecationHeaders{
		"Deprecation": fmt.Sprintf("@%d", dateToUnix(ep.DeprecatedAt)),
	}
	if ep.SunsetAt != nil {
		h["Sunset"] = dateToHTTPDate(*ep.SunsetAt)
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
