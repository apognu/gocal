package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseParameters(t *testing.T) {
	l := "HELLO;KEY1=value1;KEY2=value2"
	a, p := ParseParameters(l)

	assert.Equal(t, "HELLO", a)
	assert.Equal(t, map[string]string{"KEY1": "value1", "KEY2": "value2"}, p)
}

func Test_UnescapeString(t *testing.T) {
	l := `Hello\, world\; lorem \\ipsum.`
	l = UnescapeString(l)

	assert.Equal(t, `Hello, world; lorem \ipsum.`, l)
}

func Test_UnescapeStringWithNewline(t *testing.T) {
	l := `line1\nline2\Nline3`
	l = UnescapeString(l)

	assert.Equal(t, "line1\nline2\nline3", l)
}
