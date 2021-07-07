package mapper

import (
	"context"
	"errors"
	"log"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

var (
	ErrMultipleErrorCodeRecordsFound = errors.New("Multiple error code records found")
	ErrMultipleDatabaseRecordsFound  = errors.New("Multiple database records found")
	ErrMultipleProductRecordsFound   = errors.New("Multiple product records found")
	ErrMultipleMetadata              = errors.New("Fields received after metadata finish")
	ErrMultipleResultSetEnds         = errors.New("Got multiple result set end nodes")
	ErrFieldCountMismatch            = errors.New("Field count mismatch")
)

func (m *mapper) handleIncomingErrorCode(ctx context.Context, data fmpxmlresult.ErrorCode) error {
	if m.errorCode != nil {
		return ErrMultipleErrorCodeRecordsFound
	}

	m.errorCode = &data
	return m.flushRows(ctx)
}

func (m *mapper) handleIncomingProduct(ctx context.Context, data fmpxmlresult.Product) error {
	if m.product != nil {
		return ErrMultipleProductRecordsFound
	}

	m.product = &data
	return m.flushRows(ctx)
}

func (m *mapper) handleIncomingDatabase(ctx context.Context, data fmpxmlresult.Database) error {
	if m.database != nil {
		return ErrMultipleDatabaseRecordsFound
	}

	m.database = &data
	if err := m.flushFields(ctx); err != nil {
		return err
	}

	return m.flushRows(ctx)
}

// Here are all the handlers for the individual incoming elements
func (m *mapper) handleIncomingField(ctx context.Context, field fmpxmlresult.Field) error {
	if m.endedMetadata {
		return ErrMultipleMetadata
	}

	if !m.readyForFields() {
		m.pendingFields = append(m.pendingFields, field)
		return nil
	}

	if err := m.flushFields(ctx); err != nil {
		return err
	}

	m.fields = append(m.fields, field)

	encoder := getEncoder(field)
	joinedData := encodingFunction{
		key:   field.Name,
		proxy: encoder,
	}

	m.encodingFunctions = append(m.encodingFunctions, joinedData)
	return m.flushRows(ctx)
}

func (m *mapper) handleIncomingRow(ctx context.Context, row fmpxmlresult.NormalizedRow) error {
	if m.RowHandler == nil {
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

	if m.RowIDField != "" {
		cap++
	}

	if m.ModificationIDField != "" {
		cap++
	}

	// Pre-compute the capacity to be nicer to the garbage collector
	out.encoders = make([]encoder, 0, cap)

	if m.RowIDField != "" {
		f := func(enc *gojay.Encoder) {
			enc.StringKey(m.RowIDField, row.RecordID)
		}

		out.encoders = append(out.encoders, f)
	}

	if m.ModificationIDField != "" {
		f := func(enc *gojay.Encoder) {
			enc.StringKey(m.ModificationIDField, row.ModID)
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

	return m.RowHandler(ctx, out)
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
	return m.flushRows(ctx)
}

func (m *mapper) readyForRows() bool {
	if !m.readyForFields() {
		return false
	}

	if !m.endedMetadata {
		return false
	}

	return true
}

func (m *mapper) readyForFields() bool {
	// database is required for date/time parsing
	if m.database == nil {
		return false
	}

	return true
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

func (m *mapper) flushFields(ctx context.Context) error {
	if !m.readyForFields() {
		return nil
	}

	for _, field := range m.pendingFields {
		if err := m.handleIncomingField(ctx, field); err != nil {
			return err
		}
	}

	m.pendingFields = nil
	return nil
}
