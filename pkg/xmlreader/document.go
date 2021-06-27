package xmlreader

import (
	"encoding/xml"
	"fmt"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type document struct {
	XMLName   xml.Name  `xml:"FMPXMLRESULT"`
	ErrorCode int       `xml:"ERRORCODE"`
	Product   product   `xml:"PRODUCT"`
	Metadata  metadata  `xml:"METADATA"`
	Database  database  `xml:"DATABASE"`
	ResultSet resultSet `xml:"RESULTSET"`
}

func (d document) Normalize() (*fmpxmlresult.FMPXMLResult, error) {
	out := fmpxmlresult.FMPXMLResult{
		ErrorCode: d.ErrorCode,
	}

	var err error

	out.Product = d.Product.Normalize()

	if out.Metadata, err = d.Metadata.Normalize(); err != nil {
		return nil, fmt.Errorf("Could not parse metadata: %s", err)
	}

	if out.Database, err = d.Database.Normalize(); err != nil {
		return nil, fmt.Errorf("Could not parse database: %s", err)
	}

	if out.ResultSet, err = d.ResultSet.Normalize(); err != nil {
		return nil, fmt.Errorf("Could not parse result set: %s", err)
	}

	return &out, nil
}
