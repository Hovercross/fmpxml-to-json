package xmlreader

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type resultSet struct {
	XMLName xml.Name `xml:"RESULTSET"`
	Found   string   `xml:"FOUND,attr"`
	Rows    []row    `xml:"ROW"`
}

func (rs resultSet) Normalize() (fmpxmlresult.ResultSet, error) {
	out := fmpxmlresult.ResultSet{
		Rows: make([]fmpxmlresult.Row, len(rs.Rows)),
	}

	found, err := strconv.Atoi(rs.Found)

	if err != nil {
		return out, fmt.Errorf("Could not parse found count: '%s': %s", rs.Found, err)
	}

	out.Found = found

	for i, row := range rs.Rows {
		normalized, err := row.Normalize()

		if err != nil {
			return out, fmt.Errorf("Could not parse row %d: %s", i, err)
		}

		out.Rows[i] = normalized
	}

	return out, nil
}
