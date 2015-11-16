# gocal

Fast (and opinionated) ICAL parser in Golang.

Gocal takes an io.Reader and produces an array of ```Event```s from it.

## Usage

```go
package main

func main() {
  f, _ := os.Open("/tmp/mycalendar.ics")
  defer f.Close()

  c := gocal.NewParser(f)
  c.Parse()

  for _, e := range c.Events {
    fmt.Printf("%s on %s by %s", e.Summary, e.Start, e.Organizer.Cn)
  }
}
```

## Limitations

I do not pretend this abides by [RFC 5545](https://tools.ietf.org/html/rfc5545),
this only covers parts I needed to be parsed for my own personal use. Among
other, most propery parameters are not handled by the library, and, for now,
only the following properties are parsed:

 * ```UID```
 * ```SUMMARY``` / ```DESCRIPTION```
 * ```DTSTART``` / ```DTEND``` (day-long, local, UTC and ```TZID```d)
 * ```LOCATION```
 * ```STATUS```
 * ```ORGANIZER``` (```CN``` and value)
 * ```ATTENDEE```s (```CN```, ```PARTSTAT``` and value)
 * ```ATTACH``` (```FILENAME``` and value)

And I do not (_for now_) try and parse ```RRULE```s, so recurring events will show
as a single event.
