package parser

import (
	"strings"
)

func ParseRecurrenceParams(p string) (string, map[string]string) {
	tokens := strings.Split(p, ";")

	parameters := make(map[string]string)
	for _, p = range tokens {
		t := strings.Split(p, "=")
		if len(t) != 2 {
			continue
		}

		parameters[t[0]] = t[1]
	}

	return tokens[0], parameters
}

func ParseParameters(p string) (string, map[string]string) {
	tokens := strings.Split(p, ";")

	parameters := make(map[string]string)
	for _, p = range tokens[1:] {
		t := strings.Split(p, "=")
		if len(t) != 2 {
			continue
		}

		parameters[t[0]] = t[1]
	}

	return tokens[0], parameters
}

// Unescapes strings according to section 3.3.11 of RFC 5545
// https://tools.ietf.org/html/rfc5545#section-3.3.11
func UnescapeString(l string) string {
	l = strings.Replace(l, `\\`, `\`, -1)
	l = strings.Replace(l, `\;`, `;`, -1)
	l = strings.Replace(l, `\,`, `,`, -1)
	l = strings.Replace(l, `\n`, "\n", -1)
	l = strings.Replace(l, `\N`, "\n", -1)

	return l
}
