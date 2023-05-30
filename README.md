# gocal

Fast (and opinionated) ICAL parser in Golang.

Gocal takes an io.Reader and produces an array of `Event`s from it.

Event are parsed between two given dates (`Gocal.Start` and `Gocal.End`, 3 months by default). Any event outside this range will be ignored. This behavior can be disabled by setting `SkipBounds` to `true` in the `Gocal` struct. Please note that the behavior will still be enacted for recurring event, to prevent infinite parsing.

## Usage

```go
package main

import (
  "github.com/apognu/gocal"
)

func main() {
  f, _ := os.Open("/tmp/mycalendar.ics")
  defer f.Close()

  start, end := time.Now(), time.Now().Add(12*30*24*time.Hour)

  c := gocal.NewParser(f)
  c.Start, c.End = &start, &end
  c.Parse()

  for _, e := range c.Events {
    fmt.Printf("%s on %s by %s", e.Summary, e.Start, e.Organizer.Cn)
  }
}
```

### Timezones

Timezones specified in `TZID` attributes are parsed and expected to be parsable by Go's `time.LoadLocation()` method. If you have an ICS file using some other form of representing timezones, you can specify the mapping to be used with a callback function:

```go
var tzMapping = map[string]string{
  "My Super Zone": "Asia/Tokyo",
  "My Ultra Zone": "America/Los_Angeles",
}

gocal.SetTZMapper(func(s string) (*time.Location, error) {
  if tzid, ok := tzMapping[s]; ok {
    return time.LoadLocation(tzid)
  }
  return nil, fmt.Errorf("")
})
```

If this callback returns an `error`, the usual method of parsing the timezone will be tried. If both those methods fail, the date and time will be considered UTC.

### Custom X-* properties

Any property starting with `X-` is considered a custom property and is unmarshalled in the `event.CustomAttributes` map of string to string. For instance, a `X-LABEL` would be accessible through `event.CustomAttributes["X-LABEL"]`.

### Recurring rules

Recurring rule are automatically parsed and expanded during the period set by `Gocal.Start` and `Gocal.End`.

That being said, I try to handle the most common situations for `RRULE`s, as well as overrides (`EXDATE`s and `RECURRENCE-ID` overrides).

This was tested only lightly, I might not cover all the cases.

### Strict mode

By default, any error in parsing an event will result in the whole feed being aborted altogether (this includes missing or invalid attributes). You can change strict mode's behavior by changing the `Strict.Mode` attribute of the `Gocal` struct, with the following behavior:

 * `StrictModeFailFeed` - **default**, abort parsing of the whole feed
 * `StrictModeFailEvent` - skip the current event
 * `StrictModeFailAttribute` - skip parsing of the failing attribute, set the `Valid` attribute of the event to `false`

### Duplicate attribute behavior

The behavior when an attribute is duplicated can be customized with the `Duplicate.Mode` field. The default is to follow the configured strict mode behavior, but you can relax those rule by instructing `Gocal` to keep either the first or last value.

 * `DuplicateModeFailStrict`
 * `DuplicateModeKeepFirst`
 * `DuplicateModeKeepLast`

## Limitations

I do not pretend this abides by [RFC 5545](https://tools.ietf.org/html/rfc5545), this only covers parts I needed to be parsed for my own personal use. Among other, most property parameters are not handled by the library, and, for now, only the following properties are parsed:

 * `UID`
 * `SUMMARY` / `DESCRIPTION`
 * `DTSTART` / `DTEND` / `DURATION` (day-long, local, UTC and `TZID`d)
 * `DTSTAMP` / `CREATED` / `LAST-MODIFIED`
 * `LOCATION`
 * `STATUS`
 * `ORGANIZER` (`CN`; `DIR` and value)
 * `ATTENDEE`s (`CN`, `DIR`, `PARTSTAT` and value)
 * `ATTACH` (`FILENAME`, `ENCODING`, `VALUE`, `FMTTYPE` and value)
 * `CATEGORIES`
 * `GEO`
 * `RRULE`
 * `X-*`

Also, we ignore whatever's not a `VEVENT`.
