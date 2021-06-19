package timeconv

import (
	"bytes"
	"fmt"
	"strings"
)

// The conversion between a Filemaker format part (d) into a Golang format part
var dateConversions map[string]string = map[string]string{
	"yyyy": "2006",
	"yy":   "06",
	"mm":   "01",
	"M":    "1",
	"MM":   "01", // This is undocumented, but listed in their sample file
	"dd":   "02",
	"d":    "2",
}

// Default format splits
var dateSplits []string = []string{"-", "/", "\\", " "}

// ParseDateFormat will convert a Filemaker date format into a Go date format string
func ParseDateFormat(fmt string) string {
	// If we have a seperator, we can parse and swallow the error, since we aren't returning one.
	// If there was an error, the parsed format string will be "", which is what we would be returning anyway
	// If we care about the error, use DateSplitWithSep directly
	if sep := FindSeperator(fmt); sep != "" {
		parsed, _ := DateSplitWithSep(fmt, sep)

		return parsed
	}

	return ""
}

// Basic handling of a well-defined date format with a known seperator
func DateSplitWithSep(filemaker string, sep string) (string, error) {
	parts := strings.Split(filemaker, sep)

	// Byte buffer to keep the running format string
	buf := &bytes.Buffer{}

	for i, part := range parts {
		// Get a Go date part from a Filemaker part
		goFmt, found := dateConversions[part]

		if !found {
			return "", fmt.Errorf("Unknown part: %s", part)
		}

		// Per docstring, this cannot ever error. Freaking Go...
		buf.WriteString(goFmt)

		// Add the seperator back in as long as we aren't at the end
		if i < (len(parts) - 1) {
			buf.WriteString(sep)
		}
	}

	return buf.String(), nil
}

func FindSeperator(fmt string) string {
	// We're going to use this as our working value.
	// If we get a new candidate after this is set, it's an ambiguious format
	sep := ""

	for _, candidate := range dateSplits {
		if strings.Count(fmt, candidate) > 0 {
			// We've got a new candidate, but one was already set. No deterministic seperator.
			if sep != "" {
				return ""
			}

			sep = candidate
		}
	}

	// If we never got a candidate this will be empty, which is fine - no candidate was found
	// Duplicate candidates were handled above, so this is either the no candidate found case,
	// or the proper candidate that should have been returned
	return sep
}
