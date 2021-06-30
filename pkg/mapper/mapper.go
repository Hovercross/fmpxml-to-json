package mapper

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/stream"
	"golang.org/x/sync/errgroup"
)

type MappedRecord struct{}

var (
	ErrMultipleErrorCodeRecordsFound = errors.New("Multiple error code records found")
	ErrMultipleDatabaseRecordsFound  = errors.New("Multiple database records found")
	ErrMultipleProductRecordsFound   = errors.New("Multiple product records found")
)

func Map(ctx context.Context, r io.Reader, output chan<- MappedRecord) error {
	errGroup, ctx := errgroup.WithContext(ctx)

	errorCodeNodes := make(chan int)
	productNodes := make(chan fmpxmlresult.Product)
	fieldNodes := make(chan fmpxmlresult.Field)
	databaseNodes := make(chan fmpxmlresult.Database)
	rowNodes := make(chan fmpxmlresult.NormalizedRow)

	m := &Mapper{
		errorCodes: errorCodeNodes,
		products:   productNodes,
		fields:     fieldNodes,
		databases:  databaseNodes,
		rows:       rowNodes,

		output: output,
	}

	parser := &stream.Parser{
		Reader:     r,
		ErrorCodes: errorCodeNodes,
		Products:   productNodes,
		Fields:     fieldNodes,
		Databases:  databaseNodes,
		Rows:       rowNodes,
	}

	// Start parsing in the background, and close out all the channels after parse
	errGroup.Go(func() error {
		defer func() {
			close(errorCodeNodes)
			close(productNodes)
			close(fieldNodes)
			close(databaseNodes)
			close(rowNodes)
		}()

		return parser.Parse(ctx)
	})

	// Start reading everything in the background
	errGroup.Go(m.readErrorCodes)
	errGroup.Go(m.readProducts)
	errGroup.Go(m.readFields)
	errGroup.Go(m.readDatabases)
	errGroup.Go(m.readRows)

	return errGroup.Wait()

}

// A mapper translates the parsed rows into concrete types`
type Mapper struct {
	ErrorCode int
	Product   fmpxmlresult.Product
	Database  fmpxmlresult.Database
	Fields    []fmpxmlresult.Field

	parser stream.Parser

	output              chan<- MappedRecord
	rowIDField          string
	modificationIDField string

	m sync.RWMutex

	errorCodes <-chan int
	products   <-chan fmpxmlresult.Product
	fields     <-chan fmpxmlresult.Field
	databases  <-chan fmpxmlresult.Database
	rows       <-chan fmpxmlresult.NormalizedRow
}

func (m *Mapper) readErrorCodes() error {
	found := false

	for errorCode := range m.errorCodes {
		if found {
			return ErrMultipleErrorCodeRecordsFound
		}

		m.ErrorCode = errorCode
		found = true
	}

	return nil
}

func (m *Mapper) readProducts() error {
	found := false
	for product := range m.products {
		if found {
			return ErrMultipleProductRecordsFound
		}

		m.Product = product
		found = true
	}

	return nil
}

func (m *Mapper) readDatabases() error {
	found := false

	for database := range m.databases {
		if found {
			return ErrMultipleDatabaseRecordsFound
		}

		m.Database = database
		found = true
	}

	return nil
}

func (m *Mapper) readFields() error {
	for field := range m.fields {
		m.m.Lock()
		m.Fields = append(m.Fields, field)
		defer m.m.Unlock()
	}

	return nil
}

func (m *Mapper) readRows() error {
	for row := range m.rows {
		if err := m.readRow(row); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) readRow(row fmpxmlresult.NormalizedRow) error {
	m.m.RLock()
	defer m.m.RUnlock()

	// Normalize and do something
	return nil
}
