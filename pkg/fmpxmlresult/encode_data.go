package fmpxmlresult

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// This file has all the translation functions for strings/numbers/dates/etc to raw JSON

// a dataEncoder will convert a single string stored within a <DATA> element into an appropriate JSON value
type dataEncoder func(string) (json.RawMessage, error)

// The default string based encoder
func encodeString(s string) (json.RawMessage, error) {
	return json.Marshal(s)
}

func getTimeEncoder(inFormat, outFormat string) dataEncoder {
	out := func(s string) (json.RawMessage, error) {
		dt, err := time.Parse(inFormat, s)

		if err != nil {
			return nil, err
		}

		output := dt.Format(outFormat)

		// Swallow the error: marshaling a string can never fail
		encoded, _ := json.Marshal(output)

		return json.RawMessage(encoded), nil
	}

	return out
}

// Just check if this is a number and include it as-is. This may have an issue
// if there is a number that Go can parse, but Javascript cannot.
func passthroughEncodeNumber(s string) (json.RawMessage, error) {
	// Just attempt a float parse to see if this is numeric.
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return nil, err
	}

	// We have a parsable number. Just include it, since we don't want to lose precision doing,
	// for example, an extremely long integer to float conversion

	return json.RawMessage(s), nil
}

// Unlike the unsafe encoding, this will actually parse and then reformat the number
// to ensure it's in a friendly format - not hex or binary or anything.
func reformatEncodeNumber(s string) (json.RawMessage, error) {
	type parser func(string) (json.RawMessage, error)

	for _, f := range []parser{encodeInt, encodeFloat} {
		val, err := f(s)

		if err == nil {
			return val, nil
		}
	}

	return nil, fmt.Errorf("Could not numerically parse %s", s)
}

// int to raw JSON or error
func encodeInt(s string) (json.RawMessage, error) {
	v, err := strconv.ParseInt(s, 0, 64)

	if err != nil {
		return nil, err
	}

	return json.RawMessage(strconv.FormatInt(v, 10)), nil
}

// float to raw JSON or error
func encodeFloat(s string) (json.RawMessage, error) {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return nil, err
	}

	return json.RawMessage(fmt.Sprint(v)), nil
}
