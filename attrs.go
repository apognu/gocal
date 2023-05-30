package gocal

import (
	"fmt"
	"time"

	"github.com/apognu/gocal/parser"
)

// resolve() processes a line's value, taking duplicate rule into account to inform failure modes.
// It takes as parameters:
//   - A Gocal and Line instance
//   - A pointer to the resulting field in the in-process event
//   - A resolver function that parse the raw string value into the desired output type
//   - An optional post-processing function that can modify other event attributs, if necessary
func resolve[T comparable](gc *Gocal, l *Line, dst *T, resolve func(gc *Gocal, l *Line) (T, T, error), post func(gc *Gocal, out T)) error {
	// Retrieve the parsed value for the field, as well as its default value
	value, empty, err := resolve(gc, l)

	if err != nil {
		return err
	}

	// Apply duplicate attribute rule if the target attribute is not empty (note: would fail for empty string values)
	if dst != nil && *dst != empty {
		if gc.Duplicate.Mode == DuplicateModeFailStrict {
			return NewDuplicateAttribute(l.Key, l.Value)
		}
	}

	// If the value is empty or the duplicate mode allows further processing, set the value
	if *dst == empty || gc.Duplicate.Mode == DuplicateModeKeepLast {
		*dst = value

		if post != nil && dst != nil {
			post(gc, *dst)
		}
	}

	return nil
}

func resolveString(gc *Gocal, l *Line) (string, string, error) {
	return l.Value, "", nil
}

func resolveDate(gc *Gocal, l *Line) (*time.Time, *time.Time, error) {
	d, err := parser.ParseTime(l.Value, l.Params, parser.TimeStart, false, gc.AllDayEventsTZ)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse: %s", err)
	}

	return d, nil, nil
}

func resolveDateEnd(gc *Gocal, l *Line) (*time.Time, *time.Time, error) {
	d, err := parser.ParseTime(l.Value, l.Params, parser.TimeEnd, false, gc.AllDayEventsTZ)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse: %s", err)
	}

	return d, nil, nil
}

func resolveDuration(gc *Gocal, l *Line) (*time.Duration, *time.Duration, error) {
	d, err := parser.ParseDuration(l.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse: %s", err)
	}

	return d, nil, nil
}

func resolveOrganizer(gc *Gocal, l *Line) (*Organizer, *Organizer, error) {
	o := Organizer{
		Cn:          l.Params["CN"],
		DirectoryDn: l.Params["DIR"],
		Value:       l.Value,
	}

	return &o, nil, nil
}

func resolveGeo(gc *Gocal, l *Line) (*Geo, *Geo, error) {
	lat, long, err := parser.ParseGeo(l.Value)
	if err != nil {
		return nil, nil, err
	}

	return &Geo{lat, long}, nil, nil
}
