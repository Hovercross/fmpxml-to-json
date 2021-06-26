package fmpxmlresult

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Run through all the available number parsers to try to get a valid value out
func parseNumber(s string) (json.RawMessage, error) {
	type parser func(string) (json.RawMessage, error)

	for _, f := range []parser{parseInt, parseFloat} {
		val, err := f(s)

		if err == nil {
			return val, nil
		}
	}

	return nil, fmt.Errorf("Could not numerically parse %s", s)
}

// int to raw JSON or error
func parseInt(s string) (json.RawMessage, error) {
	v, err := strconv.ParseInt(s, 0, 64)

	if err != nil {
		return nil, err
	}

	return json.RawMessage(strconv.FormatInt(v, 10)), nil
}

// float to raw JSON or error
func parseFloat(s string) (json.RawMessage, error) {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return nil, err
	}

	return json.RawMessage(fmt.Sprint(v)), nil
}
