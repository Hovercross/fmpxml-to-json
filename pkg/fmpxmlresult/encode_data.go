package fmpxmlresult

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

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

		encoded, err := json.Marshal(output)

		if err != nil {
			return nil, err
		}

		return json.RawMessage(encoded), nil
	}

	return out
}

// Run through all the available number parsers to try to get a valid value out
func encodeNumber(s string) (json.RawMessage, error) {
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
