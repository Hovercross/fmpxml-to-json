package mapper

import (
	"errors"

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

func (m *Mapper) handleIncomingErrorCode(data int) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.result.ErrorCode != nil {
		return ErrMultipleErrorCodeRecordsFound
	}

	m.result.ErrorCode = &data
	return nil
}

func (m *Mapper) handleIncomingProduct(data fmpxmlresult.Product) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.result.Product != nil {
		return ErrMultipleProductRecordsFound
	}

	m.result.Product = &data
	return nil
}

func (m *Mapper) handleIncomingDatabase(data fmpxmlresult.Database) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.result.Database != nil {
		return ErrMultipleDatabaseRecordsFound
	}

	m.result.Database = &data
	m.tripReads()
	return nil
}

// Here are all the handlers for the individual incoming elements
func (m *Mapper) handleIncomingField(field fmpxmlresult.Field) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.gotMetadata {
		return ErrMultipleMetadata
	}

	m.result.Fields = append(m.result.Fields, field)

	encoder := getEncoder(field)
	joinedData := encodingFunction{
		key:   field.Name,
		proxy: encoder,
	}

	m.encodingFunctions = append(m.encodingFunctions, joinedData)

	return nil
}

func (m *Mapper) handleIncomingRow(row fmpxmlresult.NormalizedRow) error {
	m.m.RLock()
	defer m.m.RUnlock()

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

	m.mappedRecords <- out
	return nil
}

func (m *Mapper) handleIncomingMetadataEndSignal() error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.gotMetadata {
		return ErrMultipleMetadata
	}

	m.gotMetadata = true
	m.tripReads()

	return nil
}

func (m *Mapper) handleIncomingResultSetEndSignal() error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.gotResultSetEnd {
		return ErrMultipleResultSetEnds
	}

	m.gotResultSetEnd = true
	return nil
}
