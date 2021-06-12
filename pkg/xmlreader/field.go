package xmlreader

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlout"
)

type field struct {
	XMLName   xml.Name `xml:"FIELD"`
	EmptyOK   string   `xml:"EMPTYOK,attr"`
	MaxRepeat string   `xml:"MAXREPEAT,attr"`
	Name      string   `xml:"NAME,attr"`
	Type      string   `xml:"TYPE,attr"`
}

func (f field) Normalize() fmpxmlout.Field {
	out := fmpxmlout.Field{
		EmptyOK:   mustEmptyToBool(f.EmptyOK),
		MaxRepeat: mustParseInt(f.MaxRepeat),
		Name:      f.Name,
		Type:      f.Type,
	}

	return out
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

// This is bad design. Fix it.
func mustEmptyToBool(in string) bool {
	v, err := emptyToBool(in)

	if err != nil {
		panic(err)
	}

	return v
}

// This is bad design. Fix it.
func mustParseInt(in string) int {
	v, err := strconv.Atoi(in)

	if err != nil {
		panic(err)
	}

	return v
}
