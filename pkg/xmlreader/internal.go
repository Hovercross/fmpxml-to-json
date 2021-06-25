package xmlreader

import (
	"encoding/xml"
	"io"
	"io/ioutil"
)

func readInternalFormat(r io.Reader) (document, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return document{}, err
	}

	out := document{}
	xml.Unmarshal(data, &out)

	return out, nil
}
