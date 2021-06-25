package xmlreader_test

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/xmlreader"
)

func Test_Parse(t *testing.T) {
	sampleData := []byte(`<?xml version="1.0" encoding="UTF-8" ?>
	<FMPXMLRESULT xmlns="http://www.filemaker.com/fmpxmlresult">
		<ERRORCODE>15</ERRORCODE>
		<PRODUCT BUILD="06-07-2018" NAME="FileMaker" VERSION="Server 17.0.2"/>
		<DATABASE DATEFORMAT="M/d/yyyy" LAYOUT="Overview" NAME="test.fmp12" RECORDS="1" TIMEFORMAT="h:mm:ss a"/>
		<METADATA>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="First" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="Last" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="Email" TYPE="TEXT"/>
		</METADATA>
		<RESULTSET FOUND="1">
			<ROW MODID="196" RECORDID="683">
				<COL>
					<DATA>Adam</DATA>
				</COL>
				<COL>
					<DATA>Peacock</DATA>
				</COL>
				<COL>
					<DATA>apeacock@example.org</DATA>
				</COL>
			</ROW>
			
		</RESULTSET>
	</FMPXMLRESULT>`)

	sampleBuffer := bytes.NewBuffer(sampleData)

	parsed, err := xmlreader.ReadXML(sampleBuffer)

	if err != nil {
		t.Error(err)
		return
	}

	expectedMetadata := fmpxmlresult.Metadata{
		Fields: []fmpxmlresult.Field{
			{EmptyOK: true, MaxRepeat: 1, Name: "First", Type: "TEXT"},
			{EmptyOK: true, MaxRepeat: 1, Name: "Last", Type: "TEXT"},
			{EmptyOK: true, MaxRepeat: 1, Name: "Email", Type: "TEXT"},
		},
	}

	expectedDatabase := fmpxmlresult.Database{
		DateFormat: "M/d/yyyy",
		Layout:     "Overview",
		Name:       "test.fmp12",
		Records:    1,
		TimeFormat: "h:mm:ss a",
	}

	expectedProduct := fmpxmlresult.Product{
		Build:   "06-07-2018",
		Name:    "FileMaker",
		Version: "Server 17.0.2",
	}

	expectedResultSet := fmpxmlresult.ResultSet{
		Found: 1,
		Rows: []fmpxmlresult.Row{
			{ModID: "196", RecordID: "683", Cols: []fmpxmlresult.Col{
				{Data: []string{"Adam"}},
				{Data: []string{"Peacock"}},
				{Data: []string{"apeacock@example.org"}},
			}},
		},
	}

	expected := fmpxmlresult.FMPXMLResult{
		ErrorCode: 15,
		Product:   expectedProduct,
		Database:  expectedDatabase,
		Metadata:  expectedMetadata,
		ResultSet: expectedResultSet,
	}

	for _, diff := range deep.Equal(parsed, expected) {
		t.Error(diff)
	}
}
