package mapper

import (
	"strconv"

	"github.com/francoispqt/gojay"
)

func stringSlice(key string, vals []string) (encoder, error) {
	return func(enc *gojay.Encoder) {
		enc.AddSliceStringKey(key, vals)
	}, nil
}

// This function is used for slice string transformations - given a key, input value, and transformation function,
// it will return an encoder after applying the transformation function to all input values
func stringSliceTransform(key string, vals []string, transform func(string) (string, error)) (encoder, error) {
	transformed := make([]string, len(vals))

	// When copying from one array to another, I prefer the single return type of range and to index both values
	for i := range vals {
		var err error

		if transformed[i], err = transform(vals[i]); err != nil {
			return nil, err
		}
	}

	return stringSlice(key, transformed)
}

func (m *mapper) dateSlice(key string, vals []string) (encoder, error) {
	return stringSliceTransform(key, vals, m.reformatDate)
}

func (m *mapper) timeSlice(key string, vals []string) (encoder, error) {
	return stringSliceTransform(key, vals, m.reformatTime)
}

func (m *mapper) timestampSlice(key string, vals []string) (encoder, error) {
	return stringSliceTransform(key, vals, m.reformatDate)
}

func numericSlice(key string, vals []string) (encoder, error) {
	// We're going to encode as int64 when possible, otherwise float64
	out := make(encodableArray, len(vals))

	for i := range vals {
		var err error

		if out[i], err = numericArrayItem(vals[i]); err != nil {
			return nil, err
		}
	}

	return func(enc *gojay.Encoder) {
		enc.ArrayKey(key, out)
	}, nil
}

func rawArrayItem(s string) encodableArrayItem {
	v := encodedRawNumber(s)

	if v != nil {
		return func(enc *gojay.Encoder) {
			enc.AddEmbeddedJSON(&v)
		}
	}

	return nil
}

func intArrayItem(s string) encodableArrayItem {
	v, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return nil
	}

	return func(enc *gojay.Encoder) {
		enc.Int64(v)
	}
}

func floatArrayItem(s string) encodableArrayItem {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return nil
	}

	return func(enc *gojay.Encoder) {
		enc.Float64(v)
	}
}

func numericArrayItem(s string) (encodableArrayItem, error) {
	for _, candidate := range []func(string) encodableArrayItem{intArrayItem, floatArrayItem} {
		if encoder := candidate(s); encoder != nil {
			return encoder, nil
		}
	}

	return nil, ErrCouldNotDecodeNumber
}
