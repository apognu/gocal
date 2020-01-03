package parser

import (
	"testing"
	"time"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func Test_ParseTimeUTC(t *testing.T) {
	ti, err := ParseTime("20150910T135212Z", map[string]string{}, TimeStart)

	assert.Equal(t, nil, err)

	assert.Equal(t, 10, ti.Day())
	assert.Equal(t, time.September, ti.Month())
	assert.Equal(t, 2015, ti.Year())
	assert.Equal(t, 13, ti.Hour())
	assert.Equal(t, 52, ti.Minute())
	assert.Equal(t, 12, ti.Second())
	assert.Equal(t, time.UTC, ti.Location())
}

func Test_ParseTimezone(t *testing.T) {
	data := map[string]string{
		"Europe/Paris":         "europe/paris",
		"America/Los_Angeles":  "america/lOs_anGeles",
		"Europe/Isle_of_Man":   "europe/isle_OF_man",
		"Africa/Dar_es_Salaam": "AfricA/Dar_Es_salaam",
	}

	for exp, in := range data {
		tz, err := LoadTimezone(in)

		assert.Nil(t, err)
		assert.Equal(t, exp, tz.String())
	}
}

func Test_CustomTimezoneMapper(t *testing.T) {
	TZMapper = func(s string) (*time.Location, error) {
		mapping := map[string]string{
			"test1": "Europe/Paris",
			"test2": "America/Los_Angeles",
		}

		if tzid, ok := mapping[s]; ok {		
			return time.LoadLocation(tzid)
		}
		return nil, fmt.Errorf("mapping not found")
	}

	ti, _ := ParseTime("20150910T135212", map[string]string{"TZID": "test1"}, TimeStart)
	tz, _ := time.LoadLocation("Europe/Paris")

	assert.Equal(t, tz, ti.Location())

	ti, _ = ParseTime("20150910T135212", map[string]string{"TZID": "test2"}, TimeStart)
	tz, _ = time.LoadLocation("America/Los_Angeles")

	assert.Equal(t, tz, ti.Location())

	ti, _ = ParseTime("20150910T135212", map[string]string{"TZID": "test3"}, TimeStart)
	tz, _ = time.LoadLocation("UTC")
	
	assert.Equal(t, tz, ti.Location())
	
	ti, _ = ParseTime("20150910T135212", map[string]string{"TZID": "Europe/Paris"}, TimeStart)
	tz, _ = time.LoadLocation("Europe/Paris")

	assert.Equal(t, tz, ti.Location())

	TZMapper = nil
}

func Test_ParseTimeTZID(t *testing.T) {
	ti, err := ParseTime("20150910T135212", map[string]string{"TZID": "Europe/Paris"}, TimeStart)
	tz, _ := time.LoadLocation("Europe/Paris")

	assert.Equal(t, nil, err)

	assert.Equal(t, 10, ti.Day())
	assert.Equal(t, time.September, ti.Month())
	assert.Equal(t, 2015, ti.Year())
	assert.Equal(t, 13, ti.Hour())
	assert.Equal(t, 52, ti.Minute())
	assert.Equal(t, 12, ti.Second())
	assert.Equal(t, tz, ti.Location())
}

func Test_ParseTimeAllDayStart(t *testing.T) {
	ti, err := ParseTime("20150910", map[string]string{"VALUE": "DATE"}, TimeStart)

	assert.Equal(t, nil, err)

	assert.Equal(t, 10, ti.Day())
	assert.Equal(t, time.September, ti.Month())
	assert.Equal(t, 2015, ti.Year())
	assert.Equal(t, 0, ti.Hour())
	assert.Equal(t, 0, ti.Minute())
	assert.Equal(t, 0, ti.Second())
	assert.Equal(t, time.UTC, ti.Location())
}

func Test_ParseTimeAllDayEnd(t *testing.T) {
	ti, err := ParseTime("20150911", map[string]string{"VALUE": "DATE"}, TimeEnd)

	assert.Equal(t, nil, err)

	assert.Equal(t, 10, ti.Day())
	assert.Equal(t, time.September, ti.Month())
	assert.Equal(t, 2015, ti.Year())
	assert.Equal(t, 23, ti.Hour())
	assert.Equal(t, 59, ti.Minute())
	assert.Equal(t, 59, ti.Second())
	assert.Equal(t, time.UTC, ti.Location())
}
