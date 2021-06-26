package fmpxmlresult

import "encoding/json"

// mustMarshal changes a JSON encode error to a panic -
// used in cases where we have a defined something-or-other that can never fail a marshal, for unit test coverage
func mustMarshal(val interface{}) json.RawMessage {
	out, err := json.Marshal(val)

	if err != nil {
		panic(err)
	}

	return json.RawMessage(out)
}
