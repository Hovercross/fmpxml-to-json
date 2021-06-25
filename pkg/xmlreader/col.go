package xmlreader

import (
	"encoding/xml"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type col struct {
	XMLName xml.Name `xml:"COL"`
	Data    []string `xml:"DATA"`
}

func (c col) Normalize() fmpxmlresult.Col {
	return fmpxmlresult.Col{
		Data: c.Data,
	}
}
