package stream

import (
	"errors"

	"github.com/francoispqt/gojay"
)

var ErrMultipleValuesForScalar = errors.New("Multiple values for scalar field")

type encodable func(string, []string, *gojay.Encoder) error

func encodeString(key string, vals []string, enc *gojay.Encoder) error {
	if didNull(key, vals, enc) {
		return nil
	}

	if len(vals) != 1 {
		return ErrMultipleValuesForScalar
	}

	first := vals[0]
	enc.StringKey(key, first)
	return nil
}

func encodeStringArray(key string, vals []string, enc *gojay.Encoder) error {
	enc.SliceStringKey(key, vals)

	return nil
}

func encodeNumber(key string, vals []string, enc *gojay.Encoder) error {
	if didNull(key, vals, enc) {
		return nil
	}

	if len(vals) != 1 {
		return ErrMultipleValuesForScalar
	}

	first := vals[0]

}

func didNull(key string, vals []string, enc *gojay.Encoder) bool {
	if len(vals) == 0 {
		enc.AddNullKey(key)
		return true
	}

	return false
}
