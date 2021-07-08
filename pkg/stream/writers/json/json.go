package jsonWriter

import (
	"context"
	"encoding/json"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/mapper"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
)

type JSONResult struct {
	ErrorCode parser.ErrorCode      `json:"errorCode"`
	Database  parser.Database       `json:"database"`
	Fields    []parser.Field        `json:"fields"`
	Product   parser.Product        `json:"product"`
	Records   []mapper.MappedRecord `json:"records"`
}

func (jr *JSONResult) setDatabase(ctx context.Context, data parser.Database) error {
	jr.Database = data

	return nil
}

func (jr *JSONResult) setErrorCode(ctx context.Context, data parser.ErrorCode) error {
	jr.ErrorCode = data

	return nil
}

func (jr *JSONResult) appendField(ctx context.Context, data parser.Field) error {
	jr.Fields = append(jr.Fields, data)

	return nil
}

func (jr *JSONResult) setProduct(ctx context.Context, data parser.Product) error {
	jr.Product = data

	return nil
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

		ErrorCodeHandler: out.setErrorCode,
		DatabaseHandler:  out.setDatabase,
		FieldHandler:     out.appendField,
		ProductHandler:   out.setProduct,
	}

	err := p.Map(ctx, r)

	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
