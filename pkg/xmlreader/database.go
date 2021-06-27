package xmlreader

import (
	"fmt"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type database struct {
	DateFormat string `xml:"DATEFORMAT,attr"`
	Layout     string `xml:"LAYOUT,attr"`
	Name       string `xml:"NAME,attr"`
	Records    string `xml:"RECORDS,attr"`
	TimeFormat string `xml:"TIMEFORMAT,attr"`
}

func (d database) Normalize() (*fmpxmlresult.Database, error) {
	out := fmpxmlresult.Database{
		DateFormat: d.DateFormat,
		Layout:     d.Layout,
		Name:       d.Name,
		TimeFormat: d.TimeFormat,
	}

	records, err := strconv.Atoi(d.Records)

	if err != nil {
		return nil, fmt.Errorf("Could not parse records '%s' into integer: %s", d.Records, err)
	}

	out.Records = records

	return &out, nil
}
