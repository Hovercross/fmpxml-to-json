package mapper

import (
	"errors"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/constants"
)

var (
	ErrColumnDataMismatch   = errors.New("Multiple data values for a scalar field")
	ErrCouldNotDecodeNumber = errors.New("Number could not be decoded")
	ErrNoTimeParseLayout    = errors.New("Date/time format not defined")
)

type encoder func(*gojay.Encoder)
type encoderProxy func(string, []string) (encoder, error)
type scalarEncoderProxy func(string, string) (encoder, error)

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

func (m *mapper) getEncoder(field fmpxmlresult.Field) encoderProxy {
	// TEXT, NUMBER, DATE, TIME, TIMESTAMP, and CONTAINER
	if field.MaxRepeat == 1 {
		switch field.Type {
		case constants.NUMBER:
			return scalarWrapper(numericScalar)
		case constants.DATE:
			return scalarWrapper(m.dateScalar)
		case constants.TIME:
			return scalarWrapper(m.timeScalar)
		case constants.TIMESTAMP:
			return scalarWrapper(m.timestampScalar)
		default:
			return scalarWrapper(stringScalar)
		}
	}

	// Slices
	switch field.Type {
	case constants.NUMBER:
		// TODO
		return stringSlice
	case constants.DATE:
		return m.dateSlice
	case constants.TIME:
		return m.timeSlice
	case constants.TIMESTAMP:
		return m.timestampSlice
	default:
		return stringSlice
	}
}
