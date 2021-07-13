package mapper

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
	"go.uber.org/zap"
)

var (
	ErrMultipleErrorCodeRecordsFound = errors.New("Multiple error code records found")
	ErrMultipleDatabaseRecordsFound  = errors.New("Multiple database records found")
	ErrMultipleProductRecordsFound   = errors.New("Multiple product records found")
	ErrMultipleMetadata              = errors.New("Fields received after metadata finish")
	ErrMultipleResultSetEnds         = errors.New("Got multiple result set end nodes")
	ErrFieldCountMismatch            = errors.New("Field count mismatch")
)

func (m *mapper) handleIncomingErrorCode(ctx context.Context, log *zap.Logger, data parser.ErrorCode) error {
	log.Debug("handling incoming error codedata", zap.Int("error-code", int(data)))

	if m.gotErrorCode {
		log.Error("got duplicate error code data")

		return ErrMultipleErrorCodeRecordsFound
	}

	m.gotErrorCode = true

	if m.outgoingErrorCodes == nil {
		log.Debug("no outgoing error code handler")
		return nil
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled
	case m.outgoingErrorCodes <- data:
		log.Debug("emitted error code")
		return nil
	}
}

func (m *mapper) handleIncomingProduct(ctx context.Context, log *zap.Logger, data parser.Product) error {
	log.Debug("handling incoming product data")

	if m.gotProduct {
		log.Error("got duplicate product data")

		return ErrMultipleProductRecordsFound
	}

	m.gotProduct = true

	if m.outgoingProducts == nil {
		log.Debug("outgoing products is nil")
		return nil
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled
	case m.outgoingProducts <- data:
		log.Debug("emitted product data")
		return nil
	}
}

func (m *mapper) handleIncomingDatabase(ctx context.Context, log *zap.Logger, data parser.Database) error {
	log.Debug("handling incoming database data")

	if m.gotDatabase {
		log.Error("got duplicate database data")
		return ErrMultipleDatabaseRecordsFound
	}

	m.gotDatabase = true

	m.dateLayout = timeconv.ParseDateFormat(data.DateFormat)
	m.timeLayout = timeconv.ParseTimeFormat(data.TimeFormat)
	m.timestampLayout = m.dateLayout + " " + m.timeLayout // No idea if this is correct

	if m.outgoingDatabases == nil {
		log.Debug("no outgoing database channel")
	}

	if m.outgoingDatabases != nil {
		select {
		case <-ctx.Done():
			log.Warn("context was canceled")
			return context.Canceled
		case m.outgoingDatabases <- data:
			log.Debug("emitted database data")
		}
	}

	return m.flushRows(ctx, log)
}

// Here are all the handlers for the individual incoming elements
func (m *mapper) handleIncomingField(ctx context.Context, log *zap.Logger, field parser.Field) error {
	log.Debug("handling incoming field")

	if m.endedMetadata {
		log.Error("got field after metadata end")

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

	if m.outgoingFields == nil {
		log.Debug("outgoing field handler was nil")
		return nil
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled
	case m.outgoingFields <- field:
		log.Debug("emitted field")
		return nil
	}
}

func (m *mapper) handleIncomingRow(ctx context.Context, log *zap.Logger, row parser.NormalizedRow) error {
	log.Debug("handling incoming row")

	if m.outgoingRows == nil {
		log.Debug("outgoing row handler is nil")
		return nil
	}

	if !m.readyForRows() {
		m.pendingRows = append(m.pendingRows, row)
		log.Warn("rows not ready for processing, collecting incoming row")

		return nil
	}

	out := MappedRecord{}

	if len(row.Columns) != len(m.encodingFunctions) {
		log.Error("incoming row count mismatch",
			zap.Int("incoming-row-column-count", len(row.Columns)),
			zap.Int("encoding-function-count", len(m.encodingFunctions)))

		return ErrFieldCountMismatch
	}

	cap := len(row.Columns)

	if m.rowIDField != "" {
		cap++
	}

	if m.modificationIDField != "" {
		cap++
	}

	if m.hashField != "" {
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

	if m.hashField != "" {
		hash := sha512.New()

		for i, col := range row.Columns {
			// Write the field name, just to ensure they don't change
			hash.Write([]byte(m.encodingFunctions[i].key))

			for _, datum := range col {
				// Write the data
				hash.Write([]byte(datum))

				// Write a pad, so that []string{nil, "peacock"} has a different hash than []string{"peacock", nil}
				hash.Write([]byte("\n"))
			}
		}

		val := hex.EncodeToString(hash.Sum(nil))
		f := func(enc *gojay.Encoder) {
			enc.StringKey(m.hashField, val)
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

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled
	case m.outgoingRows <- out:
		log.Debug("emitted row code")
		return nil
	}
}

func (m *mapper) handleIncomingMetadataEndSignal(ctx context.Context, log *zap.Logger) error {
	log.Debug("handling incoming metadata end signal")

	if m.endedMetadata {
		log.Error("got duplicate metadata end signal")

		return ErrMultipleMetadata
	}

	m.endedMetadata = true

	return m.flushRows(ctx, log)
}

func (m *mapper) handleIncomingResultSetEndSignal(ctx context.Context, log *zap.Logger) error {
	log.Debug("handling incoming result set end signal")

	if m.endedResultSet {
		log.Error("got duplicate result set end signal")

		return ErrMultipleResultSetEnds
	}

	m.endedResultSet = true
	return nil
}

func (m *mapper) readyForRows() bool {
	// Metadata means we're done with fields, database means we should have date/time parsing capabilities
	return m.endedMetadata && m.gotDatabase
}

func (m *mapper) flushRows(ctx context.Context, log *zap.Logger) error {
	if !m.readyForRows() {
		return nil
	}

	for _, row := range m.pendingRows {
		log.Info("flushing held row")

		if err := m.handleIncomingRow(ctx, log, row); err != nil {
			log.Error("error when processing held row", zap.Error(err))

			return err
		}
	}

	m.pendingRows = nil
	return nil
}
