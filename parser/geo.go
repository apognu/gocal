package parser

import (
	"strconv"
	"strings"
)

func ParseGeo(l string) (float64, float64) {
	token := strings.SplitN(l, ",", 2)
	if len(token) != 2 {
		return 0.0, 0.0
	}
	lat, laterr := strconv.ParseFloat(token[0], 64)
	if laterr != nil {
		return 0.0, 0.0
	}
	long, longerr := strconv.ParseFloat(token[1], 64)
	if longerr != nil {
		return 0.0, 0.0
	}

	return lat, long
}
