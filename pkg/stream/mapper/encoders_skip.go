package mapper

import "github.com/francoispqt/gojay"

// A noop encoder, used incase we are skipping a field so that the mapper doesn't have to do a nil check
func skipEncoder(key string, vals []string) (encoder, error) {
	return func(enc *gojay.Encoder) {}, nil
}
