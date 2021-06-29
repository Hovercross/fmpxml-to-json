package timeconv

import (
	"strings"
	"time"
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

func MakeTranslationFunc(layout, format string) func(string) (string, error) {
	out := func(s string) (string, error) {
		t, err := time.Parse(layout, format)

		if err != nil {
			return "", err
		}

		return t.Format(format), nil
	}

	return out
}
