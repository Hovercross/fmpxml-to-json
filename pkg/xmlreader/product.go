package xmlreader

import "github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"

type product struct {
	Build   string `xml:"BUILD,attr"`
	Name    string `xml:"NAME,attr"`
	Version string `xml:"VERSION,attr"`
}

// Normalize to the standardized output
func (p product) Normalize() *fmpxmlresult.Product {
	return &fmpxmlresult.Product{
		Name:    p.Name,
		Build:   p.Build,
		Version: p.Version,
	}
}
