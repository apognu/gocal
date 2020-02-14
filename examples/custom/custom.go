package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/apognu/gocal"
)

const ics = `
BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTAMP:20151116T133227Z
DTSTART;TZID=Europe/Paris:20190101T090000
DTEND;TZID=Europe/Paris:20190101T110000
UID:one@gocal
SUMMARY:Event with custom labels
X-ROOMID:128-132P
X-COLOR:#000000
END:VEVENT

BEGIN:VEVENT
DTSTAMP:20151116T133227Z
DTSTART;TZID=Europe/Paris:20190201T090000
DTEND;TZID=Europe/Paris:20190201T110000
UID:two@gocal
SUMMARY:Second event with custom labels
X-ROOMID:802-127A
X-COLOR:#ffffff
END:VEVENT
END:VCALENDAR
`

func main() {
	tz, _ := time.LoadLocation("Europe/Paris")
	start, end := time.Date(1970, 1, 1, 0, 0, 0, 0, tz), time.Date(3000, 1, 1, 0, 0, 0, 0, tz)

	c := gocal.NewParser(strings.NewReader(ics))
	c.Start, c.End = &start, &end
	c.Strict = gocal.StrictParams{
		Mode: gocal.StrictModeFailAttribute,
	}
	c.Parse()

	for _, e := range c.Events {
		fmt.Printf("%s on %s - %s\n", e.Summary, e.CustomAttributes["X-ROOMID"], e.CustomAttributes["X-COLOR"])
	}
}
