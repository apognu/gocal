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
	Start          time.Time
	End            time.Time
	Location       string
	Status         string
	Organizer      Organizer
	Attendees      []Attendee
	Attachments    []Attachment
	IsRecurring    bool
	RecurrenceRule string
}

type Organizer struct {
	Cn    string
	Value string
}

type Attendee struct {
	Cn     string
	Status string
	Value  string
}

type Attachment struct {
	Filename string
	Value    string
}
