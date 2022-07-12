package gocal

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const ics = `BEGIN:VCALENDAR
METHOD:COUNTER
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
 e Popineau;X-NUM-GUESTS=0;X-RESPONSE-COMMENT="Not interested":mailto:antoi
 ne.popineau@example.net
ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;CN=John
  Connor;X-NUM-GUESTS=0:mailto:john.connor@example.net
LOCATION:My Place
COMMENT;LANGUAGE=en-US:I don't think so.
CLASS: PRIVATE
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
	start, end := time.Date(2010, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2017, 1, 1, 0, 0, 0, 0, time.Local)

	gc := NewParser(strings.NewReader(ics))
	gc.Start, gc.End = &start, &end
	gc.Parse()

	assert.Equal(t, 2, len(gc.Events))

	assert.Equal(t, "COUNTER", gc.Method)
	assert.Equal(t, "Lorem Ipsum Dolor Sit Amet", gc.Events[0].Summary)
	assert.Equal(t, "PRIVATE", gc.Events[0].Class)
	assert.Equal(t, "0001@example.net", gc.Events[0].Uid)
	assert.Equal(t, "Amazing description on two lines", gc.Events[0].Description)
	assert.Equal(t, 2, len(gc.Events[0].Attendees))
	assert.Equal(t, "Antoine Popineau", gc.Events[0].Attendees[0].Cn)
	assert.Equal(t, "0", gc.Events[0].Attendees[0].CustomAttributes["X-NUM-GUESTS"])
	assert.Equal(t, "\"Not interested\"", gc.Events[0].Attendees[0].CustomAttributes["X-RESPONSE-COMMENT"])
	assert.Equal(t, "John Connor", gc.Events[0].Attendees[1].Cn)
	assert.Equal(t, 0, len(gc.Events[0].CustomAttributes))
	assert.Equal(t, 2, len(gc.Events[1].CustomAttributes))
	assert.Equal(t, "#abc123", gc.Events[1].CustomAttributes["X-COLOR"])
}

func Test_ParseLine(t *testing.T) {
	tests := []struct {
		from         string
		expectKey    string
		expectValue  string
		expectParams map[string]string
	}{
		{
			from:         `HELLO: world`,
			expectKey:    "HELLO",
			expectValue:  "world",
			expectParams: map[string]string{},
		},
		{
			from:         `HELLO:`,
			expectKey:    "HELLO",
			expectValue:  "",
			expectParams: map[string]string{},
		},
		{
			from:         `HELLO;KEY1=value1;KEY2=value2: world`,
			expectKey:    "HELLO",
			expectValue:  "world",
			expectParams: map[string]string{"KEY1": `value1`, "KEY2": `value2`},
		},
		{
			from:         `HELLO;KEY1="foo:value1";KEY2="bar:value2": world`,
			expectKey:    "HELLO",
			expectValue:  "world",
			expectParams: map[string]string{"KEY1": `"foo:value1"`, "KEY2": `"bar:value2"`},
		},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("parse-line-%d", idx), func(t *testing.T) {
			gc := NewParser(strings.NewReader(test.from))
			gc.scanner.Scan()
			l, err, done := gc.parseLine()

			assert.Equal(t, nil, err)
			assert.Equal(t, true, done)

			assert.Equal(t, test.expectKey, l.Key)
			assert.Equal(t, test.expectValue, l.Value)
			assert.Equal(t, test.expectParams, l.Params)
		})
	}
}

func createLine(size int) string {
	return fmt.Sprintf("%s:%s", strings.Repeat("A", size), strings.Repeat("B", size))
}

func Benchmark_splitLineTokens20(b *testing.B) {
	l := createLine(10)
	for n := 0; n < b.N; n++ {
		splitLineTokens(l)
	}
}

func Benchmark_stringSplitN20(b *testing.B) {
	l := createLine(10)
	for n := 0; n < b.N; n++ {
		strings.SplitN(l, ":", 2)
	}
}

func Benchmark_splitLineTokens100(b *testing.B) {
	l := createLine(50)
	for n := 0; n < b.N; n++ {
		splitLineTokens(l)
	}
}

func Benchmark_stringSplitN100(b *testing.B) {
	l := createLine(50)
	for n := 0; n < b.N; n++ {
		strings.SplitN(l, ":", 2)
	}
}

// Event repeats every second monday and tuesday
// Instance of January, 29th is excluded
// Instance of January, 1st is changed
// Event repeats every month on the second day
const recuringICS = `BEGIN:VCALENDAR
PRODID:Gocal
VERSION:2.0
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
EXDATE;VALUE=DATE-TIME:20180129T090000Z
END:VEVENT
BEGIN:VEVENT
DTSTART:20180101T090000Z
DTEND:20180101T110000Z
DTSTAMP:20151116T133227Z
UID:0003@google.com
SUMMARY:Every two weeks on mondays and tuesdays for three events
RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU;COUNT=3
EXDATE;VALUE=DATE-TIME:20180129T090000Z
END:VEVENT
BEGIN:VEVENT
DTSTART:20180101T110000Z
DTEND:20180101T130000Z
DTSTAMP:20151116T133227Z
UID:004@google.com
RECURRENCE-ID:20180101T090000Z
SUMMARY:This changed!
END:VEVENT
END:VCALENDAR`

func Test_ReccuringRule(t *testing.T) {
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2018, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(recuringICS))
	gc.Start, gc.End = &start, &end
	gc.Parse()

	assert.Equal(t, 11, len(gc.Events))

	assert.Equal(t, "This changed!", gc.Events[0].Summary)
	assert.Equal(t, "Every month on the second", gc.Events[2].Summary)
	assert.Equal(t, "Every two weeks on mondays and tuesdays forever", gc.Events[4].Summary)
}

const recurringICSWithExdate = `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:plop
SUMMARY:Lorem ipsum dolor sit amet
DTSTAMP:20151116T133227Z
DTSTART:20190101T130000Z
DTEND:20190101T140000Z
RRULE:FREQ=MONTHLY;COUNT=5
EXDATE:20190201T130000Z
END:VEVENT
END:VCALENDAR`

func Test_ReccuringRuleWithExdate(t *testing.T) {
	start, end := time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2019, 12, 31, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(recurringICSWithExdate))
	gc.Start, gc.End = &start, &end
	gc.Parse()

	assert.Equal(t, 4, len(gc.Events))

	d := time.Date(2019, 2, 1, 13, 0, 0, 0, time.Local).Format("2006-02-01")

	for _, e := range gc.Events {
		assert.NotEqual(t, d, e.Start.Format("2016-02-01"))
	}
}

const recurringICSWithMultipleExdate = `BEGIN:VCALENDAR
PRODID:-//Google Inc//Google Calendar 70.9054//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:Family calendar
X-WR-TIMEZONE:America/New_York
X-WR-CALDESC:Esparza family events
BEGIN:VTIMEZONE
TZID:America/Grand_Turk
X-LIC-LOCATION:America/Grand_Turk
BEGIN:STANDARD
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
DTSTART:19701101T020000
RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU
END:STANDARD
BEGIN:DAYLIGHT
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
DTSTART:19700308T020000
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VTIMEZONE
TZID:America/New_York
X-LIC-LOCATION:America/New_York
BEGIN:DAYLIGHT
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
DTSTART:19700308T020000
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU
END:DAYLIGHT
BEGIN:STANDARD
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
DTSTART:19701101T020000
RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU
END:STANDARD
END:VTIMEZONE
BEGIN:VTIMEZONE
TZID:America/Phoenix
X-LIC-LOCATION:America/Phoenix
BEGIN:STANDARD
TZOFFSETFROM:-0700
TZOFFSETTO:-0700
TZNAME:MST
DTSTART:19700101T000000
END:STANDARD
END:VTIMEZONE
BEGIN:VEVENT
DTSTART;TZID=America/New_York:20201220T173000
DTEND;TZID=America/New_York:20201220T183000
EXDATE;TZID=America/New_York:20210425T173000
EXDATE;TZID=America/New_York:20211024T173000
EXDATE;TZID=America/New_York:20211031T173000
EXDATE;TZID=America/New_York:20211121T173000
EXDATE;TZID=America/New_York:20211128T173000
EXDATE;TZID=America/New_York:20220220T173000
RRULE:FREQ=WEEKLY
DTSTAMP:20220220T161319Z
UID:05F28281-077F-4059-971E-40E43F8AB3B5
URL:https://us02web.zoom.us/j/9288411040?pwd=ZVZFWVNGUWc4UHVzaHRKK010dGwrdz
	09
CREATED:20201220T170112Z
DESCRIPTION:
LAST-MODIFIED:20220205T221812Z
LOCATION:
SEQUENCE:0
STATUS:CONFIRMED
SUMMARY:Esparza family conference call 📲
TRANSP:OPAQUE
BEGIN:VALARM
ACTION:NONE
TRIGGER;VALUE=DATE-TIME:19760401T005545Z
END:VALARM
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20220218
DTEND;VALUE=DATE:20220222
DTSTAMP:20220220T161319Z
UID:6952A06D-6C4A-46D2-83DC-427A0FC5F53B
CREATED:20211021T173647Z
DESCRIPTION:
LAST-MODIFIED:20211021T173647Z
LOCATION:
SEQUENCE:0
STATUS:CONFIRMED
SUMMARY:Natalie’s Dress Shopping
TRANSP:OPAQUE
END:VEVENT
END:VCALENDAR
`

func Test_ReccuringRuleWithMultipleExdate(t *testing.T) {
	timezone := "America/New_York"
	location, err := time.LoadLocation(timezone)
	if err != nil {
		t.Errorf("Error setting timezone: %v", err)
	}

	start, end := time.Date(2022, 2, 20, 0, 0, 0, 0, location), time.Date(2022, 2, 20, 23, 59, 59, 0, location)

	gc := NewParser(strings.NewReader(recurringICSWithMultipleExdate))
	gc.Start, gc.End = &start, &end
	gc.Parse()

	assert.Equal(t, 1, len(gc.Events))
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
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2018, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(unknownICS))
	gc.Start, gc.End = &start, &end
	err := gc.Parse()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(gc.Events))
	assert.Equal(t, "Amazing description on two lines", gc.Events[0].Description)
	assert.Equal(t, "My Place", gc.Events[0].Location)
}

const invalidICS = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART;TZID=Europe/Paris:20190101T090000
DTEND;TZID=Europe/Paris:20190101T110000
UID:one@gocal
SUMMARY:Invalid event without DTSTAMP
END:VEVENT

BEGIN:VEVENT
DTSTAMP:20151116T133227Z
DTSTART;TZID=Europe/Paris:20190201T090000
DTEND;TZID=Europe/Paris:20190201T110000
UID:two@gocal
SUMMARY:Valid event
END:VEVENT
END:VCALENDAR`

func Test_InvalidEventFailFeed(t *testing.T) {
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2020, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(invalidICS))
	gc.Start, gc.End = &start, &end
	err := gc.Parse()

	assert.NotNil(t, err)
	assert.Equal(t, 0, len(gc.Events))
}

func Test_InvalidEventFailEvent(t *testing.T) {
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2020, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(invalidICS))
	gc.Start, gc.End = &start, &end
	gc.Strict = StrictParams{
		Mode: StrictModeFailEvent,
	}
	err := gc.Parse()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(gc.Events))
}

func Test_InvalidEventFailAttribute(t *testing.T) {
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2020, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(invalidICS))
	gc.Start, gc.End = &start, &end
	gc.Strict = StrictParams{
		Mode: StrictModeFailAttribute,
	}
	err := gc.Parse()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(gc.Events))
	assert.False(t, gc.Events[0].Valid)
	assert.True(t, gc.Events[1].Valid)
}

const durationICS = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTAMP:20151116T133227Z
DURATION:P1Y5DT1H10M30S
DTSTART;TZID=Europe/Paris:20190101T090000
UID:one@gocal
SUMMARY:Event with duration instead of start/end
END:VEVENT`

func Test_DurationEvent(t *testing.T) {
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2025, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(durationICS))
	gc.Start, gc.End = &start, &end
	err := gc.Parse()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(gc.Events))

	if len(gc.Events) == 1 {
		assert.Equal(t, gc.Events[0].End.Year(), 2020)
		assert.Equal(t, gc.Events[0].End.Day(), 6)
		assert.Equal(t, gc.Events[0].End.Hour(), 10)
		assert.Equal(t, gc.Events[0].End.Minute(), 10)
		assert.Equal(t, gc.Events[0].End.Second(), 30)
	}
}

const dateICS = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTAMP:20151116T133227Z
DTSTART;VALUE=DATE:20190101
DTEND;VALUE=DATE:20190101
UID:one@gocal
SUMMARY:Event with inclusive same day event
END:VEVENT
BEGIN:VEVENT
DTSTAMP:20151116T133227Z
DTSTART;VALUE=DATE:20190101
DTEND;VALUE=DATE:20190103
UID:two@gocal
SUMMARY:Event with exclusive same day event
END:VEVENT`

func Test_DateEvent(t *testing.T) {
	start, end := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2025, 2, 5, 23, 59, 59, 0, time.Local)

	gc := NewParser(strings.NewReader(dateICS))
	gc.Start, gc.End = &start, &end
	err := gc.Parse()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(gc.Events))

	if len(gc.Events) == 2 {
		assert.Equal(t, gc.Events[0].End.Year(), 2019)
		assert.Equal(t, gc.Events[0].End.Month(), time.January)
		assert.Equal(t, gc.Events[0].End.Day(), 1)
		assert.Equal(t, gc.Events[0].End.Hour(), 23)
		assert.Equal(t, gc.Events[0].End.Minute(), 59)
		assert.Equal(t, gc.Events[0].End.Second(), 59)

		assert.Equal(t, gc.Events[1].End.Year(), 2019)
		assert.Equal(t, gc.Events[1].End.Month(), time.January)
		assert.Equal(t, gc.Events[1].End.Day(), 2)
		assert.Equal(t, gc.Events[1].End.Hour(), 23)
		assert.Equal(t, gc.Events[1].End.Minute(), 59)
		assert.Equal(t, gc.Events[1].End.Second(), 59)
	}
}
