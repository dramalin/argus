package config

import (
	"time"
)

// LoadLocation returns a *time.Location for the given timezone string, or UTC if invalid.
func LoadLocation(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}
