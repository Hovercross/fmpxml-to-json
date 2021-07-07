package mapper

import (
	"context"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type CollectedData struct {
	ErrorCode fmpxmlresult.ErrorCode
	Product   fmpxmlresult.Product
	Fields    []fmpxmlresult.Field
	Database  fmpxmlresult.Database
}

type Mapper struct {
	RowIDField          string
	ModificationIDField string
	RowHandler          func(context.Context, MappedRecord) error
}

func (p Mapper) Map(ctx context.Context, r io.Reader) (CollectedData, error) {
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
