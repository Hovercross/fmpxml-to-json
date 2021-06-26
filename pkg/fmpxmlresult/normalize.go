package fmpxmlresult

import (
	"encoding/json"
	"fmt"

	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
)

// a datumNormalizer will convert a single string stored within a <DATA> element into an appropriate JSON value
type datumNormalizer func(string) (json.RawMessage, error)

// a field normalizer will take an entire column, which may or may not have repeating <DATA> elements,
// and put it into a single JSON value or an array, based on the MAXREPEAT values
type fieldNormalizer func([]string) (json.RawMessage, error)

// PopulateRecords will create all the easy to read record data
func (fmp FMPXMLResult) PopulateRecords() error {
	dateFormat := timeconv.ParseDateFormat(fmp.Database.DateFormat)
	timeFormat := timeconv.ParseTimeFormat(fmp.Database.TimeFormat)
	timestampFormat := dateFormat + " " + timeFormat // No idea if this is right, don't have an example handy

	// The specific datum normalizers we will be using for this file
	datumNormalizers := map[string]datumNormalizer{
		"DATE":      getDtNormalizer(dateFormat, "2006-01-02"),
		"TIME":      getDtNormalizer(timeFormat, "15:04:05"),
		"TIMESTAMP": getDtNormalizer(timestampFormat, "2006-01-02T15:04:05"),
		"NUMBER":    parseNumber,
	}

	// Since the fields are just in order, get the normalizers per ordered field, so we can
	// then loop over the columns in each row and apply the ordered normalizer
	normalizersByPosition := make([]fieldNormalizer, len(fmp.Metadata.Fields))

	// Load each of the normalizers
	for i, field := range fmp.Metadata.Fields {
		normalizersByPosition[i] = fmp.getNormalizer(field, datumNormalizers)
	}

	return nil
}

func (fmp FMPXMLResult) getNormalizer(f Field, normalizers map[string]datumNormalizer) fieldNormalizer {
	var dn datumNormalizer = encodeString // Just a default

	// Override the string encoder, if we have a more appropriate encoder
	if v, found := normalizers[f.Type]; found {
		dn = v
	}

	if f.MaxRepeat == 1 {
		return getSingleEncoder(dn)
	}

	return getArrayEncoder(dn)
}

// getSingleEncoder will wrap an individual datum normalizer into a
// field normalizer that doesn't do any array wrapping, but does length checks
func getSingleEncoder(f datumNormalizer) fieldNormalizer {
	// Inner function: Checks the input length, performs the parse, and then returns the result
	out := func(input []string) (json.RawMessage, error) {
		if len(input) == 0 {
			return nil, nil
		}

		if len(input) != 1 {
			return nil, fmt.Errorf("Wrong data length: got %d, wanted 1", len(input))
		}

		// Grab the single input, since this is a single encoder and we already did a length check
		input0 := input[0]

		parsed, err := f(input0)
		if err != nil {
			return nil, fmt.Errorf("Could not parse '%s': %v", input0, err)
		}

		return parsed, nil
	}

	return out
}

// getarrayEncoder will wrap an individual datum normalizer into a field normalizer that does array wrapping
func getArrayEncoder(f datumNormalizer) fieldNormalizer {
	// Inner function: Performs parses and then returns the result, along with an error if applicable
	outFunc := func(input []string) (json.RawMessage, error) {
		out := make([]json.RawMessage, len(input))

		for i, val := range input {
			// Use the normalizer to get the encoded value
			encoded, err := f(val)

			if err != nil {
				return nil, fmt.Errorf("Could not parse '%s': %v", val, err)
			}

			// Shove the pre-encoded value back into the output array
			out[i] = encoded
		}

		return json.Marshal(out)
	}

	return outFunc
}
