package gocal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apognu/gocal/parser"
)

const YmdHis = "2006-01-02 15:04:05"

func (gc *Gocal) ExpandRecurringEvent(buf *Event) []Event {
	freq := buf.RecurrenceRule["FREQ"]

	until, err := parser.ParseTime(buf.RecurrenceRule["UNTIL"], map[string]string{}, parser.TimeEnd)
	hasUntil := err == nil

	count, err := strconv.Atoi(buf.RecurrenceRule["COUNT"])
	if err != nil {
		count = -1
	}

	interval, err := strconv.Atoi(buf.RecurrenceRule["INTERVAL"])
	if err != nil {
		interval = 1
	}

	byMonth, err := strconv.Atoi(buf.RecurrenceRule["BYMONTH"])
	hasByMonth := err == nil

	byDay, ok := buf.RecurrenceRule["BYDAY"]
	hasByDay := ok

	var years, days, months int
	switch freq {
	case "DAILY":
		days = interval
		months = 0
		years = 0
		break
	case "WEEKLY":
		days = 7 * interval
		months = 0
		years = 0
		break
	case "MONTHLY":
		days = 0
		months = interval
		years = 0
		break
	case "YEARLY":
		days = 0
		months = 0
		years = interval
		break
	default:
		return []Event{}
	}

	currentCount := 0
	freqDateStart := buf.Start
	freqDateEnd := buf.End

	ev := make([]Event, 0)
	for {
		weekDaysStart := freqDateStart
		weekDaysEnd := freqDateEnd

		if !hasByMonth || strings.Contains(fmt.Sprintf("%d", byMonth), weekDaysStart.Format("1")) {
			if hasByDay {
				for i := 0; i < 7; i++ {
					excluded := false
					for _, ex := range buf.ExcludeDates {
						if ex.Equal(*weekDaysStart) {
							excluded = true
							break
						}
					}

					if !excluded {
						day := parseDayNameToIcsName(weekDaysStart.Format("Mon"))

						if strings.Contains(byDay, day) {
							currentCount++
							count--

							e := *buf
							e.Start = weekDaysStart
							e.End = weekDaysEnd
							e.Uid = buf.Uid
							e.Sequence = currentCount
							if !hasUntil || (hasUntil && until.Format(YmdHis) >= weekDaysStart.Format(YmdHis)) {
								if gc.IsInRange(e) {
									ev = append(ev, e)
								}
							}

						}
					}

					newStart := weekDaysStart.AddDate(0, 0, 1)
					newEnd := weekDaysEnd.AddDate(0, 0, 1)
					weekDaysStart = &newStart
					weekDaysEnd = &newEnd
				}
			} else {
				excluded := false
				for _, ex := range buf.ExcludeDates {
					if ex.Equal(*weekDaysStart) {
						excluded = true
						break
					}
				}

				currentCount++
				count--

				if !excluded {
					e := *buf
					e.Start = weekDaysStart
					e.End = weekDaysEnd
					e.Uid = buf.Uid
					e.Sequence = currentCount
					if !hasUntil || (hasUntil && until.Format(YmdHis) >= weekDaysStart.Format(YmdHis)) {
						if gc.IsInRange(e) {
							ev = append(ev, e)
						}
					}
				}
			}
		}

		newStart := freqDateStart.AddDate(years, months, days)
		newEnd := freqDateEnd.AddDate(years, months, days)

		freqDateStart = &newStart
		freqDateEnd = &newEnd

		if (count < 0 && weekDaysStart.After(*gc.End)) || count == 0 {
			break
		}
		if hasUntil && until.Format(YmdHis) <= freqDateStart.Format(YmdHis) {
			break
		}
	}

	return ev
}

func parseDayNameToIcsName(day string) string {
	switch day {
	case "Mon":
		return "MO"
	case "Tue":
		return "TU"
	case "Wed":
		return "WE"
	case "Thu":
		return "TH"
	case "Fri":
		return "FR"
	case "Sat":
		return "SA"
	case "Sun":
		return "SU"
	default:
		return ""
	}
}
