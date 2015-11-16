package gocal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

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
			continue
		}

		if l.Key == "BEGIN" && l.Value == "VEVENT" {
			gc.buffer = &Event{}
		} else if l.Key == "END" && l.Value == "VEVENT" {
			gc.Events = append(gc.Events, *gc.buffer)
		} else {
			gc.parseEvent(l)
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
		return nil, fmt.Errorf(""), done
	}

	attr, params := parser.ParseParameters(tokens[0])

	return &Line{Key: attr, Params: params, Value: parser.UnescapeString(strings.TrimPrefix(tokens[1], " "))}, nil, done
}

func (gc *Gocal) parseEvent(l *Line) {
	var err error

	// If this is nil, that means we did not get a BEGIN:VEVENT
	if gc.buffer == nil {
		return
	}

	switch l.Key {
	case "UID":
		gc.buffer.Uid = l.Value
	case "SUMMARY":
		gc.buffer.Summary = l.Value
	case "DESCRIPTION":
		gc.buffer.Description = l.Value
	case "DTSTART":
		gc.buffer.Start, err = parser.ParseTime(l.Value, l.Params, parser.TimeStart)
		if err != nil {
			return
		}
	case "DTEND":
		gc.buffer.End, err = parser.ParseTime(l.Value, l.Params, parser.TimeEnd)
		if err != nil {
			return
		}
	case "RRULE":
		gc.buffer.IsRecurring = true
		gc.buffer.RecurrenceRule = l.Value
	case "LOCATION":
		gc.buffer.Location = l.Value
	case "STATUS":
		gc.buffer.Status = l.Value
	case "ORGANIZER":
		gc.buffer.Organizer = Organizer{
			Cn:    l.Params["CN"],
			Value: l.Value,
		}
	case "ATTENDEE":
		gc.buffer.Attendees = append(gc.buffer.Attendees, Attendee{
			Cn:     l.Params["CN"],
			Status: l.Params["PARTSTAT"],
			Value:  l.Value,
		})
	case "ATTACH":
		gc.buffer.Attachments = append(gc.buffer.Attachments, Attachment{
			Filename: l.Params["FILENAME"],
			Value:    l.Value,
		})
	}
}
