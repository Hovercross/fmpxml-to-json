package mapper

import (
	"context"
	"errors"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
	"go.uber.org/zap"
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
	hashField           string

	fieldNameMap map[string]string

	incomingRows         chan parser.NormalizedRow
	incomingErrorCodes   chan parser.ErrorCode
	incomingProducts     chan parser.Product
	incomingFields       chan parser.Field
	incomingDatabases    chan parser.Database
	incomingMetadataEnd  chan struct{}
	incomingResultSetEnd chan struct{}

	outgoingRows       chan<- MappedRecord
	outgoingErrorCodes chan<- parser.ErrorCode
	outgoingProducts   chan<- parser.Product
	outgoingFields     chan<- parser.Field
	outgoingDatabases  chan<- parser.Database
}

type encodingFunction struct {
	key   string
	proxy encoderProxy
}

func (m *mapper) runParser(ctx context.Context, log *zap.Logger, r io.Reader) error {
	defer func() {
		close(m.incomingRows)
		close(m.incomingErrorCodes)
		close(m.incomingProducts)
		close(m.incomingFields)
		close(m.incomingDatabases)
		close(m.incomingMetadataEnd)
		close(m.incomingResultSetEnd)
	}()

	p := parser.Public{
		Rows:         m.incomingRows,
		ErrorCodes:   m.incomingErrorCodes,
		Products:     m.incomingProducts,
		Fields:       m.incomingFields,
		Databases:    m.incomingDatabases,
		MetadataEnd:  m.incomingMetadataEnd,
		ResultSetEnd: m.incomingResultSetEnd,
	}

	return p.Parse(ctx, log, r)
}

func (m *mapper) readIncomingData(ctx context.Context, log *zap.Logger) error {

	// When dealing with a for/select group of channels that will be closing down, set the channel to nil to never read from it again.
	// If the channel is left as closed, you will always get a default value out

	for {
		// All channels have been closed successfully - stop the loop
		if m.incomingRows == nil && m.incomingErrorCodes == nil && m.incomingProducts == nil && m.incomingFields == nil && m.incomingDatabases == nil && m.incomingMetadataEnd == nil && m.incomingResultSetEnd == nil {
			return nil
		}

		select {
		case incomingRow, more := <-m.incomingRows:
			if !more {
				m.incomingRows = nil
				continue
			}

			if err := m.handleIncomingRow(ctx, log, incomingRow); err != nil {
				log.Error("error when handling incoming row", zap.Error(err))
				return err
			}
		case errorCode, more := <-m.incomingErrorCodes:
			if !more {
				m.incomingErrorCodes = nil
				continue
			}

			if err := m.handleIncomingErrorCode(ctx, log, errorCode); err != nil {
				return err
			}
		case database, more := <-m.incomingDatabases:
			if !more {
				m.incomingDatabases = nil
				continue
			}

			if err := m.handleIncomingDatabase(ctx, log, database); err != nil {
				return err
			}
		case product, more := <-m.incomingProducts:
			if !more {
				m.incomingProducts = nil
				continue
			}

			if err := m.handleIncomingProduct(ctx, log, product); err != nil {
				return err
			}
		case field, more := <-m.incomingFields:
			if !more {
				m.incomingFields = nil
				continue
			}

			if err := m.handleIncomingField(ctx, log, field); err != nil {
				return err
			}
		case _, more := <-m.incomingMetadataEnd:
			if !more {
				m.incomingMetadataEnd = nil
				continue
			}

			if err := m.handleIncomingMetadataEndSignal(ctx, log); err != nil {
				return err
			}
		case _, more := <-m.incomingResultSetEnd:
			if !more {
				m.incomingResultSetEnd = nil
				continue
			}

			if err := m.handleIncomingResultSetEndSignal(ctx, log); err != nil {
				return err
			}
		}

	}
}

func (m *mapper) finalize(ctx context.Context, log *zap.Logger) error {
	return nil
}
