package fmpxmlresult

import (
	"encoding/json"
	"time"
)

func getDtNormalizer(inFormat, outFormat string) datumNormalizer {
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
