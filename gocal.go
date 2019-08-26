package gocal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/apognu/gocal/parser"
)

func NewParser(r io.Reader) *Gocal {
	return &Gocal{
		scanner: bufio.NewScanner(r),
		Events:  make([]Event, 0),
	}
}

func (gc *Gocal) Parse() error {
	if gc.Start == nil {
		start := time.Now().Add(-1 * 24 * time.Hour)
		gc.Start = &start
	}
	if gc.End == nil {
		end := time.Now().Add(3 * 30 * 24 * time.Hour)
		gc.End = &end
	}

	gc.scanner.Scan()

	rInstances := make([]Event, 0)
	ctx := &Context{Value: ContextRoot}
	for {
		l, err, done := gc.parseLine()
		if err != nil {
			if done {
				break
			}
			continue
		}

		if l.IsValue("VCALENDAR") {
			continue
		}

		if ctx.Value == ContextRoot && l.Is("BEGIN", "VEVENT") {
			ctx = ctx.Nest(ContextEvent)

			gc.buffer = &Event{}
		} else if ctx.Value == ContextEvent && l.Is("END", "VEVENT") {
			if ctx.Previous == nil {
				return fmt.Errorf("got an END:* without matching BEGIN:*")
			}
			ctx = ctx.Previous

			err := gc.checkEvent()
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("gocal error: %s", err))
			}

			if gc.buffer.IsRecurring {
				rInstances = append(rInstances, gc.ExpandRecurringEvent(gc.buffer)...)
			} else {
				if gc.buffer.Start == nil {
					continue
				}
				if gc.buffer.End == nil {
					gc.buffer.End = gc.buffer.Start
				}
				if gc.buffer.End.Before(*gc.Start) || gc.buffer.Start.After(*gc.End) {
					continue
				}

				gc.Events = append(gc.Events, *gc.buffer)
			}
		} else if l.IsKey("BEGIN") {
			ctx = ctx.Nest(ContextUnknown)
		} else if l.IsKey("END") {
			if ctx.Previous == nil {
				return fmt.Errorf("got an END:%s without matching BEGIN:%s", l.Value, l.Value)
			}
			ctx = ctx.Previous
		} else if ctx.Value == ContextEvent {
			err := gc.parseEvent(l)
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("gocal error: %s", err))
			}
		} else {
			continue
		}

		if done {
			break
		}
	}

	for _, i := range rInstances {
		if !gc.IsRecurringInstanceOverriden(&i) && gc.IsInRange(i) {
			gc.Events = append(gc.Events, i)
		}
	}

	return nil
}

func (gc *Gocal) parseLine() (*Line, error, bool) {
	// Get initial current line and check if that was the last one
	l := gc.scanner.Text()
	done := !gc.scanner.Scan()

	// If not, try and figure out if value is continued on next line
	if !done {
		for strings.HasPrefix(gc.scanner.Text(), " ") {
			l = l + strings.TrimPrefix(gc.scanner.Text(), " ")

			if done = !gc.scanner.Scan(); done {
				break
			}
		}
	}

	tokens := strings.SplitN(l, ":", 2)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("could not parse item: %s", l), done
	}

	attr, params := parser.ParseParameters(tokens[0])

	return &Line{Key: attr, Params: params, Value: parser.UnescapeString(strings.TrimPrefix(tokens[1], " "))}, nil, done
}

func (gc *Gocal) parseEvent(l *Line) error {
	var err error

	// If this is nil, that means we did not get a BEGIN:VEVENT
	if gc.buffer == nil {
		return nil
	}

	switch l.Key {
	case "UID":
		if gc.buffer.Uid != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Uid = l.Value
	case "SUMMARY":
		if gc.buffer.Summary != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Summary = l.Value
	case "DESCRIPTION":
		if gc.buffer.Description != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Description = l.Value
	case "DTSTART":
		if gc.buffer.Start != nil {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Start, err = parser.ParseTime(l.Value, l.Params, parser.TimeStart)
		gc.buffer.StartString = l.Value
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", l.Key, l.Value)
		}
	case "DTEND":
		gc.buffer.End, err = parser.ParseTime(l.Value, l.Params, parser.TimeEnd)
		gc.buffer.EndString = l.Value
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", l.Key, l.Value)
		}
	case "DTSTAMP":
		gc.buffer.Stamp, err = parser.ParseTime(l.Value, l.Params, parser.TimeStart)
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", l.Key, l.Value)
		}
	case "CREATED":
		if gc.buffer.Created != nil {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Created, err = parser.ParseTime(l.Value, l.Params, parser.TimeStart)
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", l.Key, l.Value)
		}
	case "LAST-MODIFIED":
		if gc.buffer.LastModified != nil {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.LastModified, err = parser.ParseTime(l.Value, l.Params, parser.TimeStart)
		if err != nil {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}
	case "RRULE":
		if len(gc.buffer.RecurrenceRule) != 0 {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.IsRecurring = true
		gc.buffer.RecurrenceRule, err = parser.ParseRecurrenceRule(l.Value)
	case "RECURRENCE-ID":
		if gc.buffer.RecurrenceID != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.RecurrenceID = l.Value
	case "EXDATE":
		d, err := parser.ParseTime(l.Value, map[string]string{}, parser.TimeStart)
		if err == nil {
			gc.buffer.ExcludeDates = append(gc.buffer.ExcludeDates, *d)
		}
	case "SEQUENCE":
		gc.buffer.Sequence, _ = strconv.Atoi(l.Value)
	case "LOCATION":
		if gc.buffer.Location != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Location = l.Value
	case "STATUS":
		if gc.buffer.Status != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Status = l.Value
	case "ORGANIZER":
		if gc.buffer.Organizer != nil {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.Organizer = &Organizer{
			Cn:          l.Params["CN"],
			DirectoryDn: l.Params["DIR"],
			Value:       l.Value,
		}
	case "ATTENDEE":
		gc.buffer.Attendees = append(gc.buffer.Attendees, Attendee{
			Cn:          l.Params["CN"],
			DirectoryDn: l.Params["DIR"],
			Status:      l.Params["PARTSTAT"],
			Value:       l.Value,
		})
	case "ATTACH":
		gc.buffer.Attachments = append(gc.buffer.Attachments, Attachment{
			Type:     l.Params["VALUE"],
			Encoding: l.Params["ENCODING"],
			Mime:     l.Params["FMTTYPE"],
			Filename: l.Params["FILENAME"],
			Value:    l.Value,
		})
	case "GEO":
		if gc.buffer.Geo != nil {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		lat, long, err := parser.ParseGeo(l.Value)
		if err != nil {
			return err
		}
		gc.buffer.Geo = &Geo{lat, long}
	case "CATEGORIES":
		gc.buffer.Categories = strings.Split(l.Value, ",")
	case "URL":
		gc.buffer.URL = l.Value
	default:
		key := strings.ToUpper(l.Key)
		if strings.HasPrefix(key, "X-") {
			if gc.buffer.CustomAttributes == nil {
				gc.buffer.CustomAttributes = make(map[string]string)
			}
			gc.buffer.CustomAttributes[key] = l.Value
		}
	}

	return nil
}

func (gc *Gocal) checkEvent() error {
	if gc.buffer.Uid == "" {
		return fmt.Errorf("could not parse event without UID")
	}
	if gc.buffer.Start == nil {
		return fmt.Errorf("could not parse event without DTSTART")
	}
	if gc.buffer.Stamp == nil {
		return fmt.Errorf("could not parse event without DTSTAMP")
	}

	return nil
}

func SetTZMapper(cb func(s string) (*time.Location, error)) {
	parser.TZMapper = cb
}
