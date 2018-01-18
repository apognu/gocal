package parser

import (
	"strings"
	"time"
)

const (
	TimeStart = iota
	TimeEnd
)

func ParseTime(s string, params map[string]string, ty int) (*time.Time, error) {
	var err error
	var tz *time.Location
	format := ""

	if params["VALUE"] == "DATE" || len(s) == 8 {
		t, err := time.Parse("20060102", s)
		if ty == TimeStart {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		} else if ty == TimeEnd {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Add(-1 * time.Second)
		}

		return &t, err
	}

	if strings.HasSuffix(s, "Z") {
		// If string end in 'Z', timezone is UTC
		format = "20060102T150405Z"
		tz, _ = time.LoadLocation("UTC")
	} else if params["TZID"] != "" {
		// If TZID param is given, parse in the timezone unless it is not valid
		format = "20060102T150405"
		tz, err = time.LoadLocation(strings.Title(strings.ToLower(params["TZID"])))
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
