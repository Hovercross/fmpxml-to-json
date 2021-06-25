package xmlreader

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type field struct {
	XMLName   xml.Name `xml:"FIELD"`
	EmptyOK   string   `xml:"EMPTYOK,attr"`
	MaxRepeat string   `xml:"MAXREPEAT,attr"`
	Name      string   `xml:"NAME,attr"`
	Type      string   `xml:"TYPE,attr"`
}

func (f field) Normalize() (fmpxmlresult.Field, error) {
	out := fmpxmlresult.Field{
		Name: f.Name,
		Type: f.Type,
	}

	emptyOK, err := emptyToBool(f.EmptyOK)

	if err != nil {
		return out, err
	}

	out.EmptyOK = emptyOK

	maxRepeat, err := strconv.Atoi(f.MaxRepeat)

	if err != nil {
		return out, err
	}

	out.MaxRepeat = maxRepeat

	return out, nil
}

func emptyToBool(in string) (bool, error) {
	if in == "YES" {
		return true, nil
	}

	if in == "NO" {
		return false, nil
	}

	return false, fmt.Errorf("Could not interpret %s as a bool", in)
}
