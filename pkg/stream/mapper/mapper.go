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

func (m Mapper) Map(ctx context.Context, r io.Reader) error {
	privateMapper := m.getPrivate()

	return privateMapper.Map(ctx, r)
}

func (m Mapper) getPrivate() *mapper {
	return &mapper{
		RowIDField:          m.RowIDField,
		ModificationIDField: m.ModificationIDField,

		RowHandler:       m.RowHandler,
		ErrorCodeHandler: m.ErrorCodeHandler,
		ProductHandler:   m.ProductHandler,
		FieldHandler:     m.FieldHandler,
		DatabaseHandler:  m.DatabaseHandler,
	}
}
