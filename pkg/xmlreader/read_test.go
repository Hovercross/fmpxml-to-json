package xmlreader_test

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlout"
	"github.com/hovercross/fmpxml-to-json/pkg/xmlreader"
)

func getSampleData() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8" ?>
	<FMPXMLRESULT xmlns="http://www.filemaker.com/fmpxmlresult">
		<ERRORCODE>15</ERRORCODE>
		<PRODUCT BUILD="06-07-2018" NAME="FileMaker" VERSION="Server 17.0.2"/>
		<DATABASE DATEFORMAT="M/d/yyyy" LAYOUT="Overview" NAME="ksTEACHERS.fmp12" RECORDS="397" TIMEFORMAT="h:mm:ss a"/>
		<METADATA>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="NameFirst" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="NameLast" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="NamePrefix" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="EmailSchool" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="Active Employee" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="NameUnique" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="IDTEACHER" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="Enrichment Meeting Room" TYPE="TEXT"/>
			<FIELD EMPTYOK="YES" MAXREPEAT="1" NAME="Enrichment_Description" TYPE="TEXT"/>
		</METADATA>
		<RESULTSET FOUND="397">
			<ROW MODID="196" RECORDID="683">
				<COL>
					<DATA>John</DATA>
				</COL>
				<COL>
					<DATA>Smith</DATA>
				</COL>
				<COL>
					<DATA>Mrs.</DATA>
				</COL>
				<COL>
					<DATA>jsmith@example.org</DATA>
				</COL>
				<COL>
					<DATA>Maybe</DATA>
				</COL>
				<COL>
					<DATA>SmithJ</DATA>
				</COL>
				<COL>
					<DATA>T001</DATA>
				</COL>
				<COL>
					<DATA>123</DATA>
				</COL>
				<COL>
					<DATA>Thingydo</DATA>
				</COL>
			</ROW>
			
		</RESULTSET>
	</FMPXMLRESULT>`)
}

func getSampleByteBuffer() *bytes.Buffer {
	return bytes.NewBuffer(getSampleData())
}

func Test_Parse(t *testing.T) {
	sampleData := getSampleByteBuffer()
	parsed, err := xmlreader.ReadXML(sampleData)

	if err != nil {
		t.Error(err)
		return
	}

	expectedDatabase := fmpxmlout.Database{
		DateFormat: "M/d/yyyy",
		Layout:     "Overview",
		Name:       "ksTEACHERS.fmp12",
		Records:    397,
		TimeFormat: "h:mm:ss a",
	}

	expectedProduct := fmpxmlout.Product{
		Build:   "06-07-2018",
		Name:    "FileMaker",
		Version: "Server 17.0.2",
	}

	expectedMetadata := fmpxmlout.Metadata{
		Database:  expectedDatabase,
		Product:   expectedProduct,
		ErrorCode: 15,
	}

	expectedRecord := fmpxmlout.Record{}

	expectedRecords := []fmpxmlout.Record{expectedRecord}

	expected := fmpxmlout.Document{
		Metadata: expectedMetadata,
		Records:  expectedRecords,
	}

	for _, diff := range deep.Equal(parsed, expected) {
		t.Error(diff)
	}
}
