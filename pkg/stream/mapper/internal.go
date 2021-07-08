package mapper

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
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

	errorCode *fmpxmlresult.ErrorCode
	product   *fmpxmlresult.Product
	fields    []fmpxmlresult.Field
	database  *fmpxmlresult.Database

	dateLayout      string
	timeLayout      string
	timestampLayout string

	pendingRows   []fmpxmlresult.NormalizedRow
	pendingFields []fmpxmlresult.Field

	endedMetadata  bool
	endedResultSet bool

	RowIDField          string
	ModificationIDField string

	RowHandler func(context.Context, MappedRecord) error
}

type encodingFunction struct {
	key   string
	proxy encoderProxy
}

func (m *mapper) Map(ctx context.Context, r io.Reader) (CollectedData, error) {
	var err error
	var wg sync.WaitGroup
	var out CollectedData

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
		return out, err
	}

	if err := m.flushRows(ctx); err != nil {
		return out, err
	}

	// This shouldn't happen if we got all the relevant field types
	if !m.readyForRows() {
		return out, ErrNeverReady
	}

	// Technically these errors can be ignored
	if m.errorCode == nil {
		return out, ErrNoErrorCodeRecordsFound
	}

	if m.database == nil {
		return out, ErrNoDatabaseRecordsFound
	}

	if !m.endedMetadata {
		return out, ErrNoMetadata
	}

	if !m.endedResultSet {
		return out, ErrNoResultSetEnds
	}

	if m.product == nil {
		return out, ErrNoProductRecordsFound
	}

	out.ErrorCode = *m.errorCode
	out.Database = *m.database
	out.Fields = m.fields
	out.Product = *m.product

	return out, nil
}
