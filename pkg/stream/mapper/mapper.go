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

	RowHandler       func(context.Context, MappedRecord) error
	ErrorCodeHandler func(context.Context, parser.ErrorCode) error
	ProductHandler   func(context.Context, parser.Product) error
	FieldHandler     func(context.Context, parser.Field) error
	DatabaseHandler  func(context.Context, parser.Database) error
}

func (m Mapper) Map(ctx context.Context, r io.Reader) error {
	privateMapper := m.getPrivate()

	return privateMapper.execute(ctx, r)
}

func (m Mapper) getPrivate() *mapper {
	return &mapper{
		rowIDField:          m.RowIDField,
		modificationIDField: m.ModificationIDField,

		rowHandler:       m.RowHandler,
		errorCodeHandler: m.ErrorCodeHandler,
		productHandler:   m.ProductHandler,
		fieldHandler:     m.FieldHandler,
		databaseHandler:  m.DatabaseHandler,
	}
}
