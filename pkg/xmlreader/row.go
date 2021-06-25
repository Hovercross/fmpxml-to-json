package xmlreader

import (
	"encoding/xml"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type row struct {
	XMLName  xml.Name `xml:"ROW"`
	ModID    string   `xml:"MODID,attr"`
	RecordID string   `xml:"RECORDID,attr"`
	Cols     []col    `xml:"COL"`
}

func (r row) Normalize() fmpxmlresult.Row {
	out := fmpxmlresult.Row{
		ModID:    r.ModID,
		RecordID: r.RecordID,
		Cols:     make([]fmpxmlresult.Col, len(r.Cols)),
	}

	for i, col := range r.Cols {
		out.Cols[i] = col.Normalize()
	}

	return out
}
