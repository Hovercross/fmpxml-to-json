package mapper

import (
	"context"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
)

type CollectedData struct {
	ErrorCode parser.ErrorCode `json:"errorCode"`
	Product   parser.Product   `json:"product"`
	Fields    []parser.Field   `json:"fields"`
	Database  parser.Database  `json:"database"`
}

type Mapper struct {
	RowIDField          string
	ModificationIDField string

	RowHandler       func(context.Context, MappedRecord)
	ErrorCodeHandler func(context.Context, parser.ErrorCode)
	ProductHandler   func(context.Context, parser.Product)
	FieldHandler     func(context.Context, parser.Field)
	DatabaseHandler  func(context.Context, parser.Database)
}

func (p Mapper) Map(ctx context.Context, r io.Reader) error {
	m := p.getPrivate()

	return m.Map(ctx, r)
}

func (p Mapper) getPrivate() *mapper {
	return &mapper{
		RowIDField:          p.RowIDField,
		ModificationIDField: p.ModificationIDField,
		RowHandler:          p.RowHandler,
	}
}
