package gocal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/gocal/parser"
)

func NewParser(r io.Reader) *Gocal {
	return &Gocal{
		scanner: bufio.NewScanner(r),
		Events:  make([]Event, 0),
	}
}

func (gc *Gocal) Parse() {
	gc.scanner.Scan()

	for {
		l, err, done := gc.parseLine()
		if err != nil {
			logrus.Fatal(err)
		}

		if l.Key == "BEGIN" && l.Value == "VEVENT" {
			gc.buffer = &Event{}
		} else if l.Key == "END" && l.Value == "VEVENT" {
			err := gc.checkEvent()
			if err != nil {
				logrus.Fatal(err)
			}

			gc.Events = append(gc.Events, *gc.buffer)
		} else {
			err := gc.parseEvent(l)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		if done {
			break
		}
	}
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
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", l.Key, l.Value)
		}
	case "DTEND":
		gc.buffer.End, err = parser.ParseTime(l.Value, l.Params, parser.TimeEnd)
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
		if gc.buffer.RecurrenceRule != "" {
			return fmt.Errorf("could not parse duplicate %s: %s", l.Key, l.Value)
		}

		gc.buffer.IsRecurring = true
		gc.buffer.RecurrenceRule = l.Value
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
