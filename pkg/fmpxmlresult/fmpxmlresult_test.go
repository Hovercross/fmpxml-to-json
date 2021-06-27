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
			{EmptyOK: true, MaxRepeat: 1, Name: "Birthday", Type: "DATE"},
			{EmptyOK: true, MaxRepeat: 1, Name: "Favorite Time", Type: "TIME"},
			{EmptyOK: true, MaxRepeat: 2, Name: "Favorite Number", Type: "NUMBER"},
			{EmptyOK: true, MaxRepeat: 1, Name: "Favorite Pie", Type: "TEXT"},
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
				{Data: []string{"1/11/1986"}},
				{Data: []string{"8:09:21 PM"}},
				{Data: []string{"42", "41.1"}},
				{Data: []string{}},
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  &database,
		Metadata:  &metadata,
		ResultSet: &resultSet,

		RecordIDField: "recordID",
		ModIDField:    "modificationID",
	}

	if err := sample.PopulateRecords(); err != nil {
		t.Error(err)
		return
	}

	// Records should be quoted as raw messages
	expectedRecords := []fmpxmlresult.Record{
		{
			"First":           json.RawMessage(`"Adam"`),
			"Last":            json.RawMessage(`"Peacock"`),
			"Email":           json.RawMessage(`["apeacock@example.org","apeacock-test@example.org"]`),
			"Birthday":        json.RawMessage(`"1986-01-11"`),
			"Favorite Time":   json.RawMessage(`"20:09:21"`),
			"Favorite Number": json.RawMessage(`[42,41.1]`),
			"Favorite Pie":    json.RawMessage("null"),

			"recordID":       json.RawMessage(`"683"`),
			"modificationID": json.RawMessage(`"196"`),
		},
	}

	for _, diff := range deep.Equal(sample.Records, expectedRecords) {
		t.Error(diff)
	}
}

func Test_InvalidNumber(t *testing.T) {
	metadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 1, Name: "First", Type: "NUMBER"},
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
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  &database,
		Metadata:  &metadata,
		ResultSet: &resultSet,
	}

	if err := sample.PopulateRecords(); err == nil {
		t.Error("Err is nil")

	}
}

func Test_InvalidDate(t *testing.T) {
	metadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 1, Name: "First", Type: "DATE"},
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
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  &database,
		Metadata:  &metadata,
		ResultSet: &resultSet,
	}

	if err := sample.PopulateRecords(); err == nil {
		t.Error("Err is nil")

	}
}

func Test_TooManyScalars(t *testing.T) {
	metadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 1, Name: "First", Type: "NUMBER"},
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
				{Data: []string{"42", "43", "44"}},
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  &database,
		Metadata:  &metadata,
		ResultSet: &resultSet,
	}

	if err := sample.PopulateRecords(); err == nil {
		t.Error("Err is nil")

	}
}

func Test_InvalidArray(t *testing.T) {
	metadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 2, Name: "First", Type: "NUMBER"},
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
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  &database,
		Metadata:  &metadata,
		ResultSet: &resultSet,
	}

	if err := sample.PopulateRecords(); err == nil {
		t.Error("Err is nil")
	}
}

func Test_ColumnMismatch(t *testing.T) {
	metadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 1, Name: "First", Type: "TEXT"},
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
			}},
		},
	}

	sample := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Database:  &database,
		Metadata:  &metadata,
		ResultSet: &resultSet,
	}

	if err := sample.PopulateRecords(); err == nil {
		t.Error("Err is nil")
	}
}
