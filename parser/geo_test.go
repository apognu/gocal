package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseGeo(t *testing.T) {
	lat, long := ParseGeo("32.745,128.45")

	assert.Equal(t, 32.745, lat)
	assert.Equal(t, 128.45, long)
}

func Test_ParseGeoError(t *testing.T) {
	lat, long := ParseGeo("hello,128.45")

	assert.Equal(t, 0.0, lat)
	assert.Equal(t, 0.0, long)

	lat, long = ParseGeo("12")

	assert.Equal(t, 0.0, lat)
	assert.Equal(t, 0.0, long)
}
