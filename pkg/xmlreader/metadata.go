package xmlreader

import (
	"encoding/xml"
	"fmt"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type metadata struct {
	XMLName xml.Name `xml:"METADATA"`
	Fields  []field  `xml:"FIELD"`
}

func (m metadata) Normalize() (fmpxmlresult.Metadata, error) {
	out := fmpxmlresult.Metadata{
		Fields: make([]fmpxmlresult.Field, len(m.Fields)),
	}

	for i, field := range m.Fields {
		normalized, err := field.Normalize()

		if err != nil {
			return out, fmt.Errorf("Could not parse field %d: %s", i, err)
		}

		out.Fields[i] = normalized
	}

	return out, nil
}
