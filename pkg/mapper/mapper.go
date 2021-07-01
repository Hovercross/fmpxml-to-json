package mapper

import (
	"context"
	"io"
	"sync"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/stream"
	"golang.org/x/sync/errgroup"
)

type output struct {
	ErrorCode *int                   `json:"errorCode"`
	Product   *fmpxmlresult.Product  `json:"product"`
	Database  *fmpxmlresult.Database `json:"database"`
	Fields    []fmpxmlresult.Field   `json:"fields"`
	Records   RecordOutputChannel    `json:"records"`
}

func Map(ctx context.Context, r io.Reader, w io.Writer) error {
	// This is our big channel - the row handler in the mapper will write to it, and the JSON encoder in output will read from it
	mappedRecords := make(chan MappedRecord)

	out := &output{
		Records: mappedRecords,
	}

	// These are all the channels that get passed into the parser, and then read by their respective loop handlers
	errorCodeNodes := make(chan int)
	productNodes := make(chan fmpxmlresult.Product)
	fieldNodes := make(chan fmpxmlresult.Field)
	databaseNodes := make(chan fmpxmlresult.Database)
	rowNodes := make(chan fmpxmlresult.NormalizedRow)
	metadataEndSignals := make(chan struct{})
	resultSetEndSignals := make(chan struct{})

	// This will get closed when we have all our requisite information, such as the fields, and can start to actually handle rows
	rowProcessTrip := make(chan struct{})

	// The mapper is what will do all the obnoxious row translations, reading data from the parser
	m := &Mapper{
		incomingErrorCodes:          errorCodeNodes,
		incomingProducts:            productNodes,
		incomingFields:              fieldNodes,
		incomingDatabases:           databaseNodes,
		incomingRows:                rowNodes,
		incomingMetdataEndSignals:   metadataEndSignals,
		incomingResultSetEndSignals: resultSetEndSignals,

		mappedRecords: mappedRecords,

		result:              out,
		startProcessingRows: rowProcessTrip,
	}

	// The parser will be emitting lightly normalized rows, metadata, and the like, but does not correlate fields to record columns
	parser := &stream.Parser{
		Reader:       r,
		ErrorCodes:   errorCodeNodes,
		Products:     productNodes,
		Fields:       fieldNodes,
		Databases:    databaseNodes,
		Rows:         rowNodes,
		MetadataEnd:  metadataEndSignals,
		ResultSetEnd: resultSetEndSignals,
	}

	errGroup, ctx := errgroup.WithContext(ctx)

	// Start parsing in the background, and close out all the channels after parse so that our error group can exit
	errGroup.Go(func() error {
		// Only the parser writes to these channels, so once the parser has finished close them out
		// to let all the remaining readers of those channels finish
		defer func() {
			close(errorCodeNodes)
			close(productNodes)
			close(fieldNodes)
			close(databaseNodes)
			close(rowNodes)
			close(metadataEndSignals)
			close(resultSetEndSignals)

			close(mappedRecords)
		}()

		err := parser.Parse(ctx)

		return err
	})

	// Start reading everything in the background. These all close on their respective channel closure,
	// which will be done after the parser finishes
	errGroup.Go(m.handleIncomingErrorCodes)
	errGroup.Go(m.handleIncomingProducts)
	errGroup.Go(m.handleIncomingFields)
	errGroup.Go(m.handleIncomingDatabases)
	errGroup.Go(m.handleIncomingRows)
	errGroup.Go(m.handleIncomingMetadataEndSignals)
	errGroup.Go(m.handleIncomingResultSetEndSignals)

	enc := gojay.NewEncoder(w)
	errGroup.Go(func() error {
		return enc.Encode(out)
	})

	// Orchestration note: Once the parser finishes (or has an error), all the channels will close,
	// which will cause all these subtasks to exit, and therefore the error group to finish
	return errGroup.Wait()
}

type encodingFunction struct {
	key   string
	proxy encoderProxy
}

// A mapper translates the parsed rows into concrete types`
type Mapper struct {
	parser stream.Parser

	result        *output
	mappedRecords chan<- MappedRecord // this is the incoming half of result.Records

	rowIDField          string
	modificationIDField string

	m             sync.RWMutex
	startEncoding sync.Once // Will be used to start the encoding process once we have all the non-row data

	incomingErrorCodes          <-chan int
	incomingProducts            <-chan fmpxmlresult.Product
	incomingFields              <-chan fmpxmlresult.Field
	incomingDatabases           <-chan fmpxmlresult.Database
	incomingMetdataEndSignals   <-chan struct{}
	incomingResultSetEndSignals <-chan struct{}

	incomingRows <-chan fmpxmlresult.NormalizedRow

	gotMetadata     bool
	gotResultSetEnd bool

	startProcessingRows chan struct{} // Once we've gotten all the secondary inputs, we'll trip this flag and allow rows to be processed. Until then, incoming rows will be collected.

	encodingFunctions []encodingFunction
}

// Trip the reader if the reader should have all the data that it needs to start reading rows
func (m *Mapper) tripReads() {
	if m.result.Database == nil {
		return
	}

	// Proxy for the fields, since they get continuously appended until the metadata end
	if !m.gotMetadata {
		return
	}

	m.startEncoding.Do(func() {
		close(m.startProcessingRows)
	})
}
