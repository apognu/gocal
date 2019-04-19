package gocal

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const ics = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART;VALUE=DATE:20141217
DTEND;VALUE=DATE:20141219
DTSTAMP:20151116T133227Z
UID:0001@example.net
CREATED:20141110T150010Z
DESCRIPTION:Amazing description on t
 wo lines
LAST-MODIFIED:20141110T150010Z
ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;CN=Antoin
 e Popineau;X-NUM-GUESTS=0:mailto:antoine.popineau@example.net
ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;CN=John
  Connor;X-NUM-GUESTS=0:mailto:john.connor@example.net
LOCATION:My Place
SEQUENCE:0
STATUS:CONFIRMED
SUMMARY:Lorem Ipsum Dolor Sit Amet
TRANSP:TRANSPARENT
END:VEVENT
BEGIN:VEVENT
DTSTART:20141203T130000Z
DTEND:20141203T163000Z
DTSTAMP:20151116T133227Z
UID:0002@google.com
CREATED:20141110T145426Z
DESCRIPTION:
LAST-MODIFIED:20141110T150016Z
LOCATION:Over there
SEQUENCE:1
STATUS:CONFIRMED
SUMMARY:The quick brown fox jumps over the lazy dog
TRANSP:TRANSPARENT
X-COLOR:#abc123
X-ADDRESS:432 Main St., San Francisco
END:VEVENT`

func Test_Parse(t *testing.T) {
	gc := NewParser(strings.NewReader(ics))
	tz, _ := time.LoadLocation("Europe/Paris")
	start := time.Date(2010, 1, 1, 0, 0, 0, 0, tz)
	gc.Start = &start
	end := time.Date(2017, 1, 1, 0, 0, 0, 0, tz)
	gc.End = &end
	gc.Parse()

	assert.Equal(t, 2, len(gc.Events))

	assert.Equal(t, "Lorem Ipsum Dolor Sit Amet", gc.Events[0].Summary)
	assert.Equal(t, "0001@example.net", gc.Events[0].Uid)
	assert.Equal(t, "Amazing description on two lines", gc.Events[0].Description)
	assert.Equal(t, 2, len(gc.Events[0].Attendees))
	assert.Equal(t, "John Connor", gc.Events[0].Attendees[1].Cn)
	assert.Equal(t, 0, len(gc.Events[0].CustomAttributes))
	assert.Equal(t, 2, len(gc.Events[1].CustomAttributes))
	assert.Equal(t, "#abc123", gc.Events[1].CustomAttributes["X-COLOR"])
}

func Test_ParseLine(t *testing.T) {
	gc := NewParser(strings.NewReader("HELLO;KEY1=value1;KEY2=value2: world"))
	gc.scanner.Scan()
	l, err, done := gc.parseLine()

	assert.Equal(t, nil, err)
	assert.Equal(t, true, done)

	assert.Equal(t, "HELLO", l.Key)
	assert.Equal(t, "world", l.Value)
	assert.Equal(t, map[string]string{"KEY1": "value1", "KEY2": "value2"}, l.Params)
}

// Event repeats every second monday and tuesday
// Instance of January, 29th is excluded
// Instance of January, 1st is changed
// Event repeats every month on the second day
const recuringICS = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20180102
DTEND:20180103
DTSTAMP:20151116T133227Z
UID:0001@google.com
SUMMARY:Every month on the second
RRULE:FREQ=MONTHLY;BYMONTHDAY=2
END:VEVENT
BEGIN:VEVENT
DTSTART:20180101T090000Z
DTEND:20180101T110000Z
DTSTAMP:20151116T133227Z
UID:0002@google.com
SUMMARY:Every two weeks on mondays and tuesdays forever
RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU
EXDATE;VALUE=DATE:20180129T090000Z
END:VEVENT
BEGIN:VEVENT
DTSTART:20180101T110000Z
DTEND:20180101T130000Z
DTSTAMP:20151116T133227Z
UID:0002@google.com
RECURRENCE-ID:20180101T090000Z
SUMMARY:This changed!
END:VEVENT
END:VCALENDAR`

func Test_ReccuringRule(t *testing.T) {
	gc := NewParser(strings.NewReader(recuringICS))
	tz, _ := time.LoadLocation("Europe/Paris")
	start := time.Date(2018, 1, 1, 0, 0, 0, 0, tz)
	gc.Start = &start
	end := time.Date(2018, 2, 5, 23, 59, 59, 0, tz)
	gc.End = &end
	gc.Parse()

	assert.Equal(t, 7, len(gc.Events))

	assert.Equal(t, "This changed!", gc.Events[0].Summary)
	assert.Equal(t, "Every month on the second", gc.Events[2].Summary)
	assert.Equal(t, "Every two weeks on mondays and tuesdays forever", gc.Events[4].Summary)
}

const unknownICS = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART;VALUE=DATE:20180117
DTEND;VALUE=DATE:20180119
DTSTAMP:20151116T133227Z
UID:0001@example.net
CREATED:20141110T150010Z
DESCRIPTION:Amazing description on t
 wo lines
LAST-MODIFIED:20141110T150010Z
ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;CN=Antoin
 e Popineau;X-NUM-GUESTS=0:mailto:antoine.popineau@example.net
ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;CN=John
	Connor;X-NUM-GUESTS=0:mailto:john.connor@example.net
BEGIN:SOMETHING
UID:0001@example.net
BEGIN:NESTED
BEGIN:AGAINNESTED
UID:0001@example.net
END:AGAINNESTED
END:NESTED
END:SOMETHING
LOCATION:My Place
SEQUENCE:0
STATUS:CONFIRMED
BEGIN:HELLOWORLD
END:HELLOWORLD
SUMMARY:Lorem Ipsum Dolor Sit Amet
TRANSP:TRANSPARENT
END:VEVENT`

func Test_UnknownBlocks(t *testing.T) {
	gc := NewParser(strings.NewReader(unknownICS))
	tz, _ := time.LoadLocation("Europe/Paris")
	start := time.Date(2018, 1, 1, 0, 0, 0, 0, tz)
	gc.Start = &start
	end := time.Date(2018, 2, 5, 23, 59, 59, 0, tz)
	gc.End = &end
	err := gc.Parse()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(gc.Events))
	assert.Equal(t, "Amazing description on two lines", gc.Events[0].Description)
	assert.Equal(t, "My Place", gc.Events[0].Location)
}

const newlineICS = `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:0001@example.net
DTSTAMP:20170419T172300Z
DTSTART;VALUE=DATE:20190419
DTEND;VALUE=DATE:20190419
DESCRIPTION:line1\nline2\Nline3
END:VEVENT
END:VCALENDAR`

func Test_Newline(t *testing.T) {
	gc := NewParser(strings.NewReader(newlineICS))
	tz, _ := time.LoadLocation("Asia/Omsk")
	start := time.Date(2019, 4, 19, 0, 0, 0, 0, tz)
	gc.Start = &start
	end := time.Date(2019, 4, 19, 23, 59, 59, 0, tz)
	gc.End = &end
	err := gc.Parse()

	assert.Nil(t, err)
	t.Log(gc.Events)
	assert.Equal(t, 1, len(gc.Events))
	assert.Equal(t, "line1\nline2\nline3", gc.Events[0].Description)
}
