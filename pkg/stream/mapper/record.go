package mapper

import (
	"errors"
	"strconv"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/constants"
)

var (
	ErrColumnDataMismatch   = errors.New("Multiple data values for a scalar field")
	ErrCouldNotDecodeNumber = errors.New("Number could not be decoded")
)

type encoder func(*gojay.Encoder)
type encoderProxy func(string, []string) (encoder, error)

// A mapped record is one that will be turned into a single row in the records array,
// indexed by field name instead of position. Our final output.
type MappedRecord struct {
	encoders []encoder
}

func (mr MappedRecord) MarshalJSONObject(enc *gojay.Encoder) {
	for _, encoder := range mr.encoders {
		encoder(enc)
	}
}

func (mr MappedRecord) IsNil() bool {
	return false
}

func (mr MappedRecord) MarshalJSON() ([]byte, error) {
	return gojay.MarshalJSONObject(mr)
}

func nullEncoder(key string, vals []string) encoder {
	if len(vals) == 0 {
		return func(enc *gojay.Encoder) {
			enc.AddNullKey(key)
		}
	}

	return nil
}

func scalarCheck(vals []string) error {
	if len(vals) != 1 {
		return ErrColumnDataMismatch
	}

	return nil
}

func (m *mapper) getEncoder(field fmpxmlresult.Field) encoderProxy {
	// TEXT, NUMBER, DATE, TIME, TIMESTAMP, and CONTAINER
	if field.MaxRepeat == 1 {
		switch field.Type {
		case constants.NUMBER:
			return getScalarNumberEncoder
		case constants.DATE:
			// TODO: Date scalar
			return stringScalar
		case constants.TIME:
			// TODO: Time scalar
			return stringScalar
		case constants.TIMESTAMP:
			// TODO: Timestamp encoder
		default:
			return stringScalar
		}
	}

	// Slices
	switch field.Type {
	case constants.NUMBER:
		// TODO
		return stringSlice
	case constants.DATE:
		// TODO
		return stringSlice
	case constants.TIME:
		// TODO
		return stringSlice
	case constants.TIMESTAMP:
		// TODO
		return stringSlice
	default:
		return stringSlice
	}
}

func getScalarNumberEncoder(key string, vals []string) (encoder, error) {
	if f := nullEncoder(key, vals); f != nil {
		return f, nil
	}

	if err := scalarCheck(vals); err != nil {
		return nil, err
	}

	val := vals[0]

	try := []func(string, string) encoder{getScalarIntEncoder, getScalarFloatEncoder}

	for _, f := range try {
		if encoder := f(key, val); encoder != nil {
			return encoder, nil
		}
	}

	return nil, ErrCouldNotDecodeNumber
}

func getScalarIntEncoder(key, val string) encoder {
	v, err := strconv.ParseInt(val, 10, 64)

	if err != nil {
		return nil
	}

	return func(enc *gojay.Encoder) {
		enc.AddInt64Key(key, v)
	}
}

func getScalarFloatEncoder(key, val string) encoder {
	v, err := strconv.ParseFloat(val, 64)

	if err != nil {
		return nil
	}

	return func(enc *gojay.Encoder) {
		enc.AddFloat64Key(key, v)
	}
}

func stringScalar(key string, vals []string) (encoder, error) {
	if f := nullEncoder(key, vals); f != nil {
		return f, nil
	}

	if err := scalarCheck(vals); err != nil {
		return nil, err
	}

	scalar := vals[0]
	return func(enc *gojay.Encoder) {
		enc.AddStringKey(key, scalar)
	}, nil
}

func stringSlice(key string, vals []string) (encoder, error) {
	return func(enc *gojay.Encoder) {
		enc.AddSliceStringKey(key, vals)
	}, nil
}
