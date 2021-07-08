package jsonWriter

import (
	"context"
	"encoding/json"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/mapper"
)

type JSONResult struct {
	ErrorCode fmpxmlresult.ErrorCode `json:"errorCode"`
	Database  fmpxmlresult.Database  `json:"database"`
	Fields    []fmpxmlresult.Field   `json:"fields"`
	Product   fmpxmlresult.Product   `json:"product"`
	Records   []mapper.MappedRecord  `json:"records"`
}

func WriteJSON(ctx context.Context, r io.Reader, w io.Writer, recordIDField, modIDField string) error {
	out := JSONResult{}

	collect := func(ctx context.Context, row mapper.MappedRecord) error {
		out.Records = append(out.Records, row)
		return nil
	}

	p := mapper.Mapper{
		RowIDField:          recordIDField,
		ModificationIDField: modIDField,
		RowHandler:          collect,
	}

	data, err := p.Map(ctx, r)

	if err != nil {
		return err
	}

	out.Database = data.Database
	out.ErrorCode = data.ErrorCode
	out.Fields = data.Fields
	out.Product = data.Product

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}