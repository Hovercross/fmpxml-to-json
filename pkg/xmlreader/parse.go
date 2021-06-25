package xmlreader

import (
	"fmt"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

func ReadXML(r io.Reader) (fmpxmlresult.FMPXMLResult, error) {
	internal, err := readInternalFormat(r)

	if err != nil {
		return fmpxmlresult.FMPXMLResult{}, fmt.Errorf("Could not read XML: %s", err)
	}

	out, err := internal.Normalize()

	if err != nil {
		return out, fmt.Errorf("Could not normalize data: %s", err)
	}

	return out, nil
}
