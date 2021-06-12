package xmlreader

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlout"
)

type product struct {
	Build   string `xml:"BUILD,attr"`
	Name    string `xml:"NAME,attr"`
	Version string `xml:"VERSION,attr"`
}

type database struct {
	DateFormat string `xml:"DATEFORMAT,attr"`
	Layout     string `xml:"LAYOUT,attr"`
	Name       string `xml:"NAME,attr"`
	Records    string `xml:"RECORDS,attr"`
	TimeFormat string `xml:"TIMEFORMAT,attr"`
}

type metadata struct {
	XMLName xml.Name `xml:"METADATA"`
	Fields  []field  `xml:"FIELD"`
}

type resultSet struct {
	XMLName xml.Name `xml:"RESULTSET"`
	Found   string   `xml:"FOUND,attr"`
	Rows    []row    `xml:"ROW"`
}

type col struct {
	XMLName xml.Name `xml:"COL"`
	Data    []string `xml:"DATA"`
}

type row struct {
	XMLName  xml.Name `xml:"ROW"`
	ModID    string   `xml:"MODID,attr"`
	RecordID string   `xml:"RECORDID,attr"`
	Cols     []col    `xml:"COL"`
}

type fmpxmlresult struct {
	XMLName   xml.Name  `xml:"FMPXMLRESULT"`
	ErrorCode int       `xml:"ERRORCODE"`
	Product   product   `xml:"PRODUCT"`
	Metadata  metadata  `xml:"METADATA"`
	Database  database  `xml:"DATABASE"`
	ResultSet resultSet `xml:"RESULTSET"`
}

func ReadXML(r io.Reader) (fmpxmlout.Document, error) {
	internal, err := readInternalFormat(r)

	if err != nil {
		return fmpxmlout.Document{}, err
	}

	return internal.ToNormalized()
}

func readInternalFormat(r io.Reader) (fmpxmlresult, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return fmpxmlresult{}, err
	}

	out := fmpxmlresult{}
	xml.Unmarshal(data, &out)

	return out, nil
}

func (o fmpxmlresult) ToNormalized() (fmpxmlout.Document, error) {
	out := fmpxmlout.Document{}

	// Code below is structured with local 'v' because Go's scopes annoy me sometimes

	// Convert database record count into int
	if v, err := strconv.Atoi(o.Database.Records); err == nil {
		out.Metadata.Database.Records = v
	} else {
		return out, err
	}

	out.Metadata.ErrorCode = o.ErrorCode

	out.Metadata.Database.DateFormat = o.Database.DateFormat
	out.Metadata.Database.Layout = o.Database.Layout
	out.Metadata.Database.Name = o.Database.Name
	out.Metadata.Database.TimeFormat = o.Database.TimeFormat

	out.Metadata.Product.Build = o.Product.Build
	out.Metadata.Product.Name = o.Product.Name
	out.Metadata.Product.Version = o.Product.Version

	out.Metadata.Fields = make([]*fmpxmlout.Field, len(o.Metadata.Fields))

	for i, f := range o.Metadata.Fields {
		field := f.Normalize()

		out.Metadata.Fields[i] = &field
	}

	out.Records = make([]fmpxmlout.Record, len(o.ResultSet.Rows))

	for rowNum, row := range o.ResultSet.Rows {
		record := fmpxmlout.Record{
			RecordID: row.RecordID,
			ModID:    row.ModID,
			Data:     map[*fmpxmlout.Field][]string{},
		}

		for colNum, col := range row.Cols {
			field := out.Metadata.Fields[colNum]

			record.Data[field] = make([]string, len(col.Data))

			for dataIndex, data := range col.Data {
				record.Data[field][dataIndex] = data
			}
		}

		out.Records[rowNum] = record

	}

	return out, nil
}
