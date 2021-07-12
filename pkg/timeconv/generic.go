package timeconv

import (
	"fmt"
	"strings"
	"time"
)

type TimeConversionError struct {
	Layout string
	Input  string
	Err    error
}

func (tce *TimeConversionError) Error() string {
	return fmt.Sprintf("Unable to parse '%s' using '%s': %v", tce.Input, tce.Layout, tce.Err)
}

func (tce *TimeConversionError) Unwrap() error {
	return tce.Err
}

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
			return "", &TimeConversionError{
				Layout: layout,
				Input:  s,
				Err:    err,
			}
		}

		return t.Format(format), nil
	}

	return out
}
