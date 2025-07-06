package models

import (
	"encoding/json"
	"time"
)

// GMTPlus7Time is a wrapper around time.Time that serializes to JSON 
// with GMT+7 timezone preserved instead of converting to UTC
type GMTPlus7Time struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface
// This preserves the GMT+7 timezone when serializing to JSON
func (t GMTPlus7Time) MarshalJSON() ([]byte, error) {
	// Format time in RFC3339 format but preserve the timezone
	stamp := t.Format(`"2006-01-02T15:04:05.999999999Z07:00"`)
	return []byte(stamp), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *GMTPlus7Time) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	// Parse the time string in RFC3339 format
	pt, err := time.Parse(`"2006-01-02T15:04:05.999999999Z07:00"`, string(b))
	if err != nil {
		return err
	}

	*t = GMTPlus7Time{pt}
	return nil
}

// NewGMTPlus7Time creates a new GMTPlus7Time from a standard time.Time
func NewGMTPlus7Time(t time.Time) GMTPlus7Time {
	// Always ensure time is in GMT+7 timezone
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return GMTPlus7Time{t.In(loc)}
}
