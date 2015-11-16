package gocal

import (
	"bufio"
	"time"
)

type Gocal struct {
	scanner *bufio.Scanner
	Events  []Event
	buffer  *Event
}

type Line struct {
	Key    string
	Params map[string]string
	Value  string
}

type Event struct {
	Uid            string
	Summary        string
	Description    string
	Categories     []string
	Start          *time.Time
	End            *time.Time
	Stamp          *time.Time
	Created        *time.Time
	LastModified   *time.Time
	Location       string
	Geo            *Geo
	Status         string
	Organizer      *Organizer
	Attendees      []Attendee
	Attachments    []Attachment
	IsRecurring    bool
	RecurrenceRule string
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
	Cn          string
	DirectoryDn string
	Status      string
	Value       string
}

type Attachment struct {
	Encoding string
	Type     string
	Mime     string
	Filename string
	Value    string
}
