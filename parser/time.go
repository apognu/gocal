package parser

import (
	"strings"
	"time"

	duration "github.com/ChannelMeter/iso8601duration"
)

const (
	TimeStart = iota
	TimeEnd
)

var (
	TZMapper func(s string) (*time.Location, error)
)

func ParseTime(s string, params map[string]string, ty int, allday bool, allDayTZ *time.Location) (*time.Time, error) {
	var err error
	var tz *time.Location

	format := ""

	if params["VALUE"] == "DATE" || len(s) == 8 {
		/*
			Reference: https://icalendar.org/iCalendar-RFC-5545/3-3-4-date.html
			DATE values are a specific format.  They should not include time information
		*/
		t, err := time.Parse("20060102", s)
		if ty == TimeStart {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, allDayTZ)
		} else if ty == TimeEnd {
			if allday {
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999, allDayTZ)
			} else {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, allDayTZ).Add(-1 * time.Millisecond)
			}
		}

		return &t, err
	}

	if strings.HasSuffix(s, "Z") {
		// If string end in 'Z', timezone is UTC
		format = "20060102T150405Z"
		tz, _ = time.LoadLocation("UTC")
	} else if params["TZID"] != "" {
		var err error

		// If TZID param is given, parse in the timezone unless it is not valid
		format = "20060102T150405"
		if TZMapper != nil {
			tz, err = TZMapper(params["TZID"])
		}
		if TZMapper == nil || err != nil {
			tz, err = LoadTimezone(params["TZID"])
		}

		if err != nil {
			tz, _ = time.LoadLocation("UTC")
		}
	} else {
		// Else, consider the timezone is local the parser
		format = "20060102T150405"
		tz = time.Local
	}

	t, err := time.ParseInLocation(format, s, tz)

	return &t, err
}

func ParseDuration(s string) (*time.Duration, error) {
	d, err := duration.FromString(s)
	if err != nil {
		return nil, err
	}
	dur := d.ToDuration()
	return &dur, nil
}

func LoadTimezone(tzid string) (*time.Location, error) {
	tz, err := time.LoadLocation(tzid)
	if err == nil {
		return tz, err
	}

	tokens := strings.Split(tzid, "_")
	for idx, t := range tokens {
		t = strings.ToLower(t)

		if t != "of" && t != "es" {
			tokens[idx] = strings.Title(t)
		} else {
			tokens[idx] = t
		}
	}

	return time.LoadLocation(strings.Join(tokens, "_"))
}
