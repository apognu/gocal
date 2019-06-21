# gocal

Fast (and opinionated) ICAL parser in Golang.

Gocal takes an io.Reader and produces an array of ```Event```s from it.

Event are parsed between two given dates (```Gocal.Start``` and ```Gocal.End```, 3 months by default). Any event outside this range will be ignored.

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

### Custom X-* properties

Any property starting with ```X-``` is considered a custom property and is unmarshalled in the ```event.CustomAttributes``` map of string to string. For instance, a ```X-LABEL``` would be accessible through ```event.CustomAttributes["X-LABEL"]```.

### Recurring rules

Recurring rule are automatically parsed and expanded during the period set by ```Gocal.Start``` and ```Gocal.End```.

That being said, I try to handle the most common situations for ```RRULE```s, as well as overrides (```EXDATE```s and ```RECURRENCE-ID``` overrides).

This was tested only lightly, I might not cover all the cases.

## Limitations

I do not pretend this abides by [RFC 5545](https://tools.ietf.org/html/rfc5545), this only covers parts I needed to be parsed for my own personal use. Among other, most property parameters are not handled by the library, and, for now, only the following properties are parsed:

 * ```UID```
 * ```SUMMARY``` / ```DESCRIPTION```
 * ```DTSTART``` / ```DTEND``` (day-long, local, UTC and ```TZID```d)
 * ```DTSTAMP``` / ```CREATED``` / ```LAST-MODIFIED```
 * ```LOCATION```
 * ```STATUS```
 * ```ORGANIZER``` (```CN```; ```DIR``` and value)
 * ```ATTENDEE```s (```CN```, ```DIR```, ```PARTSTAT``` and value)
 * ```ATTACH``` (```FILENAME```, ```ENCODING```, ```VALUE```, ```FMTTYPE``` and value)
 * ```CATEGORIES```
 * ```GEO```
 * ```RRULE```
 * ```X-*```

Also, we ignore whatever's not a ```VEVENT``` or ```VTIMEZONE```.
