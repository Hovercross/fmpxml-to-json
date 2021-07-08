package mapper

import (
	"context"
	"errors"
	"log"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
)

var (
	ErrMultipleErrorCodeRecordsFound = errors.New("Multiple error code records found")
	ErrMultipleDatabaseRecordsFound  = errors.New("Multiple database records found")
	ErrMultipleProductRecordsFound   = errors.New("Multiple product records found")
	ErrMultipleMetadata              = errors.New("Fields received after metadata finish")
	ErrMultipleResultSetEnds         = errors.New("Got multiple result set end nodes")
	ErrFieldCountMismatch            = errors.New("Field count mismatch")
)

func (m *mapper) handleIncomingErrorCode(ctx context.Context, data parser.ErrorCode) error {
	if m.gotErrorCode {
		return ErrMultipleErrorCodeRecordsFound
	}

	m.gotErrorCode = true

	if m.errorCodeHandler != nil {
		if err := m.errorCodeHandler(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

func (m *mapper) handleIncomingProduct(ctx context.Context, data parser.Product) error {
	if m.gotProduct {
		return ErrMultipleProductRecordsFound
	}

	m.gotProduct = true

	if m.productHandler != nil {
		if err := m.productHandler(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

func (m *mapper) handleIncomingDatabase(ctx context.Context, data parser.Database) error {
	if m.gotDatabase {
		return ErrMultipleDatabaseRecordsFound
	}

	m.gotDatabase = true

	m.dateLayout = timeconv.ParseDateFormat(data.DateFormat)
	m.timeLayout = timeconv.ParseTimeFormat(data.TimeFormat)
	m.timestampLayout = m.dateLayout + " " + m.timeLayout // No idea if this is correct

	if m.databaseHandler != nil {
		if err := m.databaseHandler(ctx, data); err != nil {
			return err
		}
	}

	return m.flushRows(ctx)
}

// Here are all the handlers for the individual incoming elements
func (m *mapper) handleIncomingField(ctx context.Context, field parser.Field) error {
	if m.endedMetadata {
		return ErrMultipleMetadata
	}

	// Since the date/time encoders are hung off the mapper, we don't need to worry that we don't yet have a
	// date or time format - they'll be populated before we process any rows
	encoder := m.getEncoder(field)
	joinedData := encodingFunction{
		key:   field.Name,
		proxy: encoder,
	}

	m.encodingFunctions = append(m.encodingFunctions, joinedData)

	if m.fieldHandler != nil {
		if err := m.fieldHandler(ctx, field); err != nil {
			return err
		}
	}

	return nil
}

func (m *mapper) handleIncomingRow(ctx context.Context, row parser.NormalizedRow) error {
	if m.rowHandler == nil {
		return nil
	}

	if !m.readyForRows() {
		m.pendingRows = append(m.pendingRows, row)
		log.Printf("Rows not ready for processing, collecting incoming row")
	}

	out := MappedRecord{}

	if len(row.Columns) != len(m.encodingFunctions) {
		return ErrFieldCountMismatch
	}

	cap := len(row.Columns)

	if m.rowIDField != "" {
		cap++
	}

	if m.modificationIDField != "" {
		cap++
	}

	// Pre-compute the capacity to be nicer to the garbage collector
	out.encoders = make([]encoder, 0, cap)

	if m.rowIDField != "" {
		f := func(enc *gojay.Encoder) {
			enc.StringKey(m.rowIDField, row.RecordID)
		}

		out.encoders = append(out.encoders, f)
	}

	if m.modificationIDField != "" {
		f := func(enc *gojay.Encoder) {
			enc.StringKey(m.modificationIDField, row.ModID)
		}

		out.encoders = append(out.encoders, f)
	}

	for i, proxy := range m.encodingFunctions {
		colData := row.Columns[i]

		encoder, err := proxy.proxy(proxy.key, colData)

		if err != nil {
			return err
		}

		out.encoders = append(out.encoders, encoder)
	}

	return m.rowHandler(ctx, out)
}

func (m *mapper) handleIncomingMetadataEndSignal(ctx context.Context) error {

	if m.endedMetadata {
		return ErrMultipleMetadata
	}

	m.endedMetadata = true

	return m.flushRows(ctx)
}

func (m *mapper) handleIncomingResultSetEndSignal(ctx context.Context) error {
	if m.endedResultSet {
		return ErrMultipleResultSetEnds
	}

	m.endedResultSet = true
	return nil
}

func (m *mapper) readyForRows() bool {
	// Metadata means we're done with fields, database means we should have date/time parsing capabilities
	return m.endedMetadata && m.gotDatabase
}

func (m *mapper) flushRows(ctx context.Context) error {
	if !m.readyForRows() {
		return nil
	}

	for _, row := range m.pendingRows {
		if err := m.handleIncomingRow(ctx, row); err != nil {
			return err
		}
	}

	m.pendingRows = nil
	return nil
}
