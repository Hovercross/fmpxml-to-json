package mapper

import (
	"regexp"

	"github.com/francoispqt/gojay"
)

var rawNumberRe = regexp.MustCompile(`^-?[0-9]+(.[0-9]+){0,1}$`)

// This doesn't necessarily hit 100% of numeric coverage, but it hits all the common forms.
// Definitely missed '-.3'
func encodedRawNumber(s string) gojay.EmbeddedJSON {
	if rawNumberRe.MatchString(s) {
		return gojay.EmbeddedJSON(s)
	}

	return nil
}
