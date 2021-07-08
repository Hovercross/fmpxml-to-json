package jsonWriter

import (
	"context"
	"sync"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/mapper"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/parser"
	"go.uber.org/zap"
)

type JSONResult struct {
	ErrorCode parser.ErrorCode      `json:"errorCode"`
	Database  parser.Database       `json:"database"`
	Fields    []parser.Field        `json:"fields"`
	Product   parser.Product        `json:"product"`
	Records   []mapper.MappedRecord `json:"records"`
}

func (jr *JSONResult) setDatabase(ctx context.Context, data parser.Database) error {
	jr.Database = data

	return nil
}

func (jr *JSONResult) setErrorCode(ctx context.Context, data parser.ErrorCode) error {
	jr.ErrorCode = data

	return nil
}

func (jr *JSONResult) appendField(ctx context.Context, data parser.Field) error {
	jr.Fields = append(jr.Fields, data)

	return nil
}

func (jr *JSONResult) setProduct(ctx context.Context, data parser.Product) error {
	jr.Product = data

	return nil
}

type collectorReader struct {
	result *JSONResult

	errorCodes <-chan parser.ErrorCode
	databases  <-chan parser.Database
	fields     <-chan parser.Field
	products   <-chan parser.Product
	records    <-chan mapper.MappedRecord
}

func (cr *collectorReader) run(ctx context.Context, log *zap.Logger) {
	wg := sync.WaitGroup{}

	type bgFunc func(context.Context, *zap.Logger)
	bgFuncs := []bgFunc{cr.collectErrorCodes,
		cr.collectDatabases,
		cr.collectFields,
		cr.collectProducts,
		cr.collectRecords,
	}

	// type bgFunc func(context.Context)
	for _, f := range bgFuncs {

		wg.Add(1)

		go func(f bgFunc) {
			defer wg.Done()
			f(ctx, log)
		}(f)
	}
}

func (cr *collectorReader) collectRecords(ctx context.Context, log *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			log.Warn("got context cancelation")

			return
		case v, more := <-cr.records:
			if !more {
				log.Debug("record channel closed")

				return
			}

			cr.result.Records = append(cr.result.Records, v)
		}
	}
}

func (cr *collectorReader) collectErrorCodes(ctx context.Context, log *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			log.Warn("got context cancelation")

			return
		case v, more := <-cr.errorCodes:
			if !more {
				log.Debug("error code channel closed")
				return
			}

			log.Debug("set error code", zap.Int("error-code", int(v)))
			cr.result.ErrorCode = v
		}
	}
}

func (cr *collectorReader) collectDatabases(ctx context.Context, log *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			log.Warn("got context cancelation")

			return
		case v, more := <-cr.databases:
			if !more {
				log.Debug("database channel closed")

				return
			}

			cr.result.Database = v
		}
	}
}

func (cr *collectorReader) collectFields(ctx context.Context, log *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			log.Warn("got context cancelation")

			return
		case v, more := <-cr.fields:
			if !more {
				log.Debug("field channel closed")

				return
			}

			cr.result.Fields = append(cr.result.Fields, v)
		}
	}
}

func (cr *collectorReader) collectProducts(ctx context.Context, log *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			log.Warn("got context cancelation")

			return
		case v, more := <-cr.products:
			if !more {
				log.Debug("product channel closed")

				return
			}

			cr.result.Product = v
		}
	}
}
