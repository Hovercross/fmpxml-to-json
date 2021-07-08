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

	rowIDField          string
	modificationIDField string

	rowHandler       func(context.Context, MappedRecord) error
	errorCodeHandler func(context.Context, parser.ErrorCode) error
	productHandler   func(context.Context, parser.Product) error
	fieldHandler     func(context.Context, parser.Field) error
	databaseHandler  func(context.Context, parser.Database) error
}

type encodingFunction struct {
	key   string
	proxy encoderProxy
}

func (m *mapper) execute(ctx context.Context, r io.Reader) error {
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

	if len(m.pendingRows) > 0 {
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
