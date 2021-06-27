package fmpxmlresult

import (
	"encoding/json"
	"fmt"
)

// This file has the utilities for scalar and array encoding, given a specific data encoder

// getScalarEncoder will wrap an individual datum normalizer into a
// field normalizer that doesn't do any array wrapping, but does length checks
func getScalarEncoder(f dataEncoder) fieldEncoder {
	// Inner function: Checks the input length, performs the parse, and then returns the result
	out := func(input []string) (json.RawMessage, error) {
		if len(input) == 0 {
			return json.Marshal(nil)
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
func getArrayEncoder(f dataEncoder) fieldEncoder {
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
