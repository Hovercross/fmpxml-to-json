package timeconv

import (
	"strings"
)

type conversion struct {
	from string
	to   string
}

func convert(fmt string, conversions []conversion) string {
	out := fmt

	for _, c := range conversions {
		out = strings.ReplaceAll(out, c.from, c.to)
	}

	return out
}
