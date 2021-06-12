package xmlreader

import (
	"encoding/xml"
	"io"
	"io/ioutil"
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

type field struct {
	XMLName   xml.Name `xml:"FIELD"`
	EmptyOK   string   `xml:"EMPTYOK,attr"`
	MaxRepeat string   `xml:"MAXREPEAT,attr"`
	Name      string   `xml:"NAME,attr"`
	Type      string   `xml:"TYPE,attr"`
}

type resultSet struct {
	XMLName xml.Name `xml:"RESULTSET"`
	Found   string   `xml:"FOUND,attr"`
	Rows    []row    `xml:"ROW"`
}

type col struct {
	XMLName xml.Name `xml:"COL"`
	Data    string   `xml:"DATA"`
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

func ReadXML(r io.Reader) (Document, error) {
	internal, err := readInternalFormat(r)

	if err != nil {
		return Document{}, err
	}

	return internal.ToNormalized(), nil
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

func (o fmpxmlresult) ToNormalized() Document {
	out := Document{}

	out.Metadata.ErrorCode = o.ErrorCode

	out.Metadata.Database.DateFormat = o.Database.DateFormat
	out.Metadata.Database.Layout = o.Database.Layout
	out.Metadata.Database.Name = o.Database.Name
	out.Metadata.Database.Records = o.Database.Records
	out.Metadata.Database.TimeFormat = o.Database.TimeFormat

	out.Metadata.Product.Build = o.Product.Build
	out.Metadata.Product.Name = o.Product.Name
	out.Metadata.Product.Version = o.Product.Version

	records := make([]Record, len(o.ResultSet.Rows))
	out.Records = records

	return out
}
