package gocal

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/apognu/gocal/parser"
)

const (
	StrictModeFailFeed = iota
	StrictModeFailAttribute
	StrictModeFailEvent
)

const (
	DuplicateModeFailStrict = iota
	DuplicateModeKeepFirst
	DuplicateModeKeepLast
)

type StrictParams struct {
	Mode int
}

type DuplicateParams struct {
	Mode int
}

type DuplicateAttributeError struct {
	Key, Value string
}

func NewDuplicateAttribute(k, v string) DuplicateAttributeError {
	return DuplicateAttributeError{Key: k, Value: v}
}

func (err DuplicateAttributeError) Error() string {
	return fmt.Sprintf("duplicate attribute %s: %s", err.Key, err.Value)
}

type Gocal struct {
	scanner        *bufio.Scanner
	Events         []Event
	SkipBounds     bool
	Strict         StrictParams
	Duplicate      DuplicateParams
	buffer         *Event
	Start          *time.Time
	End            *time.Time
	Method         string
	AllDayEventsTZ *time.Location
}

const (
	ContextRoot = iota
	ContextEvent
	ContextUnknown
)

type Context struct {
	Value    int
	Previous *Context
}

func (ctx *Context) Nest(value int) *Context {
	return &Context{Value: value, Previous: ctx}
}

func (gc *Gocal) IsInRange(d Event) bool {
	if (d.Start.Before(*gc.Start) && d.End.After(*gc.Start)) ||
		(d.Start.After(*gc.Start) && d.End.Before(*gc.End)) ||
		(d.Start.Before(*gc.End) && d.End.After(*gc.End)) {
		return true
	}
	return false
}

func (gc *Gocal) IsRecurringInstanceOverriden(instance *Event) bool {
	for _, e := range gc.Events {
		if e.Uid == instance.Uid {
			rid, _ := parser.ParseTime(e.RecurrenceID, e.RawStart.Params, parser.TimeStart, false, gc.AllDayEventsTZ)
			if rid.Equal(*instance.Start) {
				return true
			}
		}
	}
	return false
}

type Line struct {
	Key    string
	Params map[string]string
	Value  string
}

func (l *Line) Is(key, value string) bool {
	if strings.TrimSpace(l.Key) == key && strings.TrimSpace(l.Value) == value {
		return true
	}
	return false
}

func (l *Line) IsKey(key string) bool {
	return strings.TrimSpace(l.Key) == key
}

func (l *Line) IsValue(value string) bool {
	return strings.TrimSpace(l.Value) == value
}

type RawDate struct {
	Params map[string]string
	Value  string
}

type Event struct {
	delayed []*Line

	Uid              string
	Summary          string
	Description      string
	Categories       []string
	Start            *time.Time
	RawStart         RawDate
	End              *time.Time
	RawEnd           RawDate
	Duration         *time.Duration
	Stamp            *time.Time
	Created          *time.Time
	LastModified     *time.Time
	Location         string
	Geo              *Geo
	URL              string
	Status           string
	Organizer        *Organizer
	Attendees        []Attendee
	Attachments      []Attachment
	IsRecurring      bool
	RecurrenceID     string
	RecurrenceRule   map[string]string
	ExcludeDates     []time.Time
	Sequence         int
	CustomAttributes map[string]string
	Valid            bool
	Comment          string
	Class            string
}

type Geo struct {
	Lat  float64
	Long float64
}

type Organizer struct {
	Cn          string
	DirectoryDn string
	Value       string
}

type Attendee struct {
	Cn               string
	DirectoryDn      string
	Status           string
	Value            string
	CustomAttributes map[string]string
}

type Attachment struct {
	Encoding string
	Type     string
	Mime     string
	Filename string
	Value    string
}
