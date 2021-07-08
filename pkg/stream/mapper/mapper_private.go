package mapper

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
)

var (
	ErrNoErrorCodeRecordsFound = errors.New("No error code records found")
	ErrNoDatabaseRecordsFound  = errors.New("No database records found")
	ErrNoProductRecordsFound   = errors.New("No product records found")
	ErrNoMetadata              = errors.New("Metadata never closed")
	ErrNoResultSetEnds         = errors.New("Result set never closed")
	ErrNeverReady              = errors.New("Mapper never became ready, check for metadata and database fields")
)

// A mapper translates the parsed rows into concrete types, while emitting records and returning the collected non-record data
type mapper struct {
	encodingFunctions []encodingFunction

	dateLayout      string
	timeLayout      string
	timestampLayout string

	pendingRows   []parser.NormalizedRow
	pendingFields []parser.Field

	endedMetadata  bool
	endedResultSet bool
	gotErrorCode   bool
	gotProduct     bool
	gotDatabase    bool

	RowIDField          string
	ModificationIDField string

	RowHandler       func(context.Context, MappedRecord)
	ErrorCodeHandler func(context.Context, parser.ErrorCode)
	ProductHandler   func(context.Context, parser.Product)
	FieldHandler     func(context.Context, parser.Field)
	DatabaseHandler  func(context.Context, parser.Database)
}

type encodingFunction struct {
	key   string
	proxy encoderProxy
}

func (m *mapper) Map(ctx context.Context, r io.Reader) error {
	var err error
	var wg sync.WaitGroup

	// The parser will be emitting lightly normalized rows, metadata, and the like, but does not correlate fields to record columns
	parser := &parser.Parser{
		Reader: r,

		ErrorCodeHandler:    m.handleIncomingErrorCode,
		ProductHandler:      m.handleIncomingProduct,
		FieldHandler:        m.handleIncomingField,
		DatabaseHandler:     m.handleIncomingDatabase,
		RowHandler:          m.handleIncomingRow,
		MetadataEndHandler:  m.handleIncomingMetadataEndSignal,
		ResultSetEndHandler: m.handleIncomingResultSetEndSignal,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = parser.Parse(ctx)
	}()

	wg.Wait()

	// Parser error is priority
	if err != nil {
		return err
	}

	// Flush any pending rows - should be a noop
	if err := m.flushRows(ctx); err != nil {
		return err
	}

	// This shouldn't happen if we got all the relevant field types
	if !m.readyForRows() {
		return ErrNeverReady
	}

	// Technically these errors can be ignored
	if !m.gotErrorCode {
		return ErrNoErrorCodeRecordsFound
	}

	if !m.gotDatabase {
		return ErrNoDatabaseRecordsFound
	}

	if !m.endedMetadata {
		return ErrNoMetadata
	}

	if !m.endedResultSet {
		return ErrNoResultSetEnds
	}

	if !m.gotProduct {
		return ErrNoProductRecordsFound
	}

	return nil
}