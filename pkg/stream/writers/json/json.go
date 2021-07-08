package jsonWriter

import (
	"context"
	"encoding/json"
	"io"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/mapper"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type JSONWriter struct {
	RecordIDField string
	ModIDField    string
}

func (jw *JSONWriter) Write(ctx context.Context, log *zap.Logger, r io.Reader, w io.Writer) error {
	out := JSONResult{}
	eg, ctx := errgroup.WithContext(ctx)

	records := make(chan mapper.MappedRecord)
	errorCodes := make(chan parser.ErrorCode)
	databases := make(chan parser.Database)
	fields := make(chan parser.Field)
	products := make(chan parser.Product)

	reader := collectorReader{
		result: &out,

		records:    records,
		errorCodes: errorCodes,
		databases:  databases,
		fields:     fields,
		products:   products,
	}

	// Start up the reader
	eg.Go(func() error {
		reader.run(ctx, log)

		return nil
	})

	p := mapper.Mapper{
		RowIDField:          jw.RecordIDField,
		ModificationIDField: jw.ModIDField,

		Rows:       records,
		ErrorCodes: errorCodes,
		Databases:  databases,
		Fields:     fields,
		Products:   products,
	}

	// Start up the mapper
	eg.Go(func() error {
		// Once the mapper finishes, close out all the channels so the reader can finish
		defer func() {
			close(records)
			close(errorCodes)
			close(databases)
			close(fields)
			close(products)
		}()

		return p.Map(ctx, log, r)
	})

	// Wait for the mapper and collector to finish
	if err := eg.Wait(); err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
