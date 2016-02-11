package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseGeo(t *testing.T) {
	lat, long, err := ParseGeo("32.745;128.45")

	assert.Equal(t, nil, err)

	assert.Equal(t, 32.745, lat)
	assert.Equal(t, 128.45, long)
}

func Test_ParseGeoError(t *testing.T) {
	lat, long, err := ParseGeo("hello;128.45")

	assert.NotEqual(t, nil, err)

	assert.Equal(t, 0.0, lat)
	assert.Equal(t, 0.0, long)

	lat, long, err = ParseGeo("12")

	assert.NotEqual(t, nil, err)

	assert.Equal(t, 0.0, lat)
	assert.Equal(t, 0.0, long)
}
