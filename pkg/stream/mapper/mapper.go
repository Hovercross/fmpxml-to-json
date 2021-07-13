package mapper

import (
	"context"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
	HashField           string

	Rows       chan<- MappedRecord
	ErrorCodes chan<- parser.ErrorCode
	Products   chan<- parser.Product
	Fields     chan<- parser.Field
	Databases  chan<- parser.Database
}

func (m Mapper) Map(ctx context.Context, log *zap.Logger, r io.Reader) error {
	private := m.copy()
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return private.readIncomingData(ctx, log)
	})

	eg.Go(func() error {
		return private.runParser(ctx, log, r)
	})

	return eg.Wait()
}

func (m Mapper) copy() *mapper {
	return &mapper{
		rowIDField:          m.RowIDField,
		modificationIDField: m.ModificationIDField,
		hashField:           m.HashField,

		incomingRows:         make(chan parser.NormalizedRow),
		incomingErrorCodes:   make(chan parser.ErrorCode),
		incomingProducts:     make(chan parser.Product),
		incomingFields:       make(chan parser.Field),
		incomingDatabases:    make(chan parser.Database),
		incomingMetadataEnd:  make(chan struct{}),
		incomingResultSetEnd: make(chan struct{}),

		outgoingRows:       m.Rows,
		outgoingErrorCodes: m.ErrorCodes,
		outgoingProducts:   m.Products,
		outgoingFields:     m.Fields,
		outgoingDatabases:  m.Databases,
	}
}
