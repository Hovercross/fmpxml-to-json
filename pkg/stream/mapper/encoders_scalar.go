package mapper

import (
	"strconv"

	"github.com/francoispqt/gojay"
)

// Given that evert set of values comes in as an array, this will transform
// the slice encoder for one that is expecting a scalar, after doing null
// and multiple value checks
func scalarWrapper(f scalarEncoderProxy) encoderProxy {
	out := func(key string, vals []string) (encoder, error) {
		if len(vals) == 0 {
			return func(enc *gojay.Encoder) {
				enc.AddNullKey(key)
			}, nil
		}

		if len(vals) != 1 {
			return nil, ErrColumnDataMismatch
		}

		return f(key, vals[0])
	}

	return out
}

func stringScalar(key string, val string) (encoder, error) {
	return func(enc *gojay.Encoder) {
		enc.AddStringKey(key, val)
	}, nil
}

// This function is used for scalar string transformations - given a key, input value, and transformation function,
// it will return an encoder after applying the transformation function
func stringScalarTransform(key, val string, transform func(string) (string, error)) (encoder, error) {
	transformed, err := transform(val)

	if err != nil {
		return nil, err
	}

	return stringScalar(key, transformed)
}

func (m *mapper) dateScalar(key string, val string) (encoder, error) {
	return stringScalarTransform(key, val, m.reformatDate)
}

func (m *mapper) timeScalar(key string, val string) (encoder, error) {
	return stringScalarTransform(key, val, m.reformatTime)
}

func (m *mapper) timestampScalar(key string, val string) (encoder, error) {
	return stringScalarTransform(key, val, m.reformatTimestamp)
}

// Since Filemaker doesn't define an int vs a float, we're going to try both. If we can use an int, use an int since it has
// a higher range of precision integer values when compared with a float.
func numericScalar(key string, val string) (encoder, error) {
	try := []func(string, string) encoder{rawNumericScalar, intScalar, floatScalar}

	for _, f := range try {
		if encoder := f(key, val); encoder != nil {
			return encoder, nil
		}
	}

	return nil, ErrCouldNotDecodeNumber
}

// See if we can do this with no parsing
func rawNumericScalar(key, val string) encoder {
	if raw := encodedRawNumber(val); raw != nil {
		return func(enc *gojay.Encoder) {
			enc.AddEmbeddedJSONKey(key, &raw)
		}
	}

	return nil
}

func intScalar(key, val string) encoder {
	v, err := strconv.ParseInt(val, 10, 64)

	if err != nil {
		return nil
	}

	return func(enc *gojay.Encoder) {
		enc.AddInt64Key(key, v)
	}
}

func floatScalar(key, val string) encoder {
	v, err := strconv.ParseFloat(val, 64)

	if err != nil {
		return nil
	}

	return func(enc *gojay.Encoder) {
		enc.AddFloat64Key(key, v)
	}
}
