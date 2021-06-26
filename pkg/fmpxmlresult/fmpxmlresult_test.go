package fmpxmlresult_test

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

func Test_Populate(t *testing.T) {
	metadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 1, Name: "First", Type: "TEXT"},
			{EmptyOK: true, MaxRepeat: 1, Name: "Last", Type: "TEXT"},
			{EmptyOK: false, MaxRepeat: 2, Name: "Email", Type: "TEXT"},
		},
	}

	database := fmpxmlresult.Database{
		DateFormat: "M/d/yyyy",
		Layout:     "Overview",
		Name:       "test.fmp12",
		Records:    1,
		TimeFormat: "h:mm:ss a",
	}

	resultSet := fmpxmlresult.ResultSet{
		Found: 1,
		Rows: []fmpxmlresult.Row{
			{ModID: "196", RecordID: "683", Cols: []fmpxmlresult.Col{
				{Data: []string{"Adam"}},
				{Data: []string{"Peacock"}},
				{Data: []string{"apeacock@example.org", "apeacock-test@example.org"}},
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  database,
		Metadata:  metadata,
		ResultSet: resultSet,
	}

	if err := sample.PopulateRecords(); err != nil {
		t.Error(err)
	}

	// Records should be quoted as raw messages
	expectedRecords := []fmpxmlresult.Record{
		{
			"First": json.RawMessage(`"Adam"`),
			"Last":  json.RawMessage(`"Peacock"`),
			"Email": json.RawMessage(`"apeacock@example.org","apeacock-test@example.org"`),
		},
	}

	for _, diff := range deep.Equal(sample.Records, expectedRecords) {
		t.Error(diff)
	}
}
