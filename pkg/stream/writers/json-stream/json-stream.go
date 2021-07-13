package jsonStreamWrite

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/mapper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	ErrJSONTooLong = errors.New("When writing fixed-width length data, the JSON record would not fit inside said size")
)

type StreamWriter struct {
	RecordIDField string
	ModIDField    string
	HashField     string
	Prefix        string
	Suffix        string
	LengthSize    int
}

func (sr StreamWriter) getRowWriter(w io.Writer, ch <-chan mapper.MappedRecord) *mappedRecordWriter {
	return &mappedRecordWriter{
		Prefix:     sr.Prefix,
		Suffix:     sr.Suffix,
		LengthSize: sr.LengthSize,

		w:  bufio.NewWriter(w),
		ch: ch,
	}
}

func (sr StreamWriter) Write(ctx context.Context, log *zap.Logger, r io.Reader, w io.Writer) error {
	ch := make(chan mapper.MappedRecord)

	rowWriter := sr.getRowWriter(w, ch)

	eg, ctx := errgroup.WithContext(ctx)

	mapper := mapper.Mapper{
		RowIDField:          sr.RecordIDField,
		ModificationIDField: sr.ModIDField,
		HashField:           sr.HashField,

		Rows: ch,
	}

	eg.Go(func() error {
		defer func() {
			close(ch)
		}()

		log.Debug("starting mapper")
		return mapper.Map(ctx, log, r)
	})

	eg.Go(func() error {
		log.Debug("starting writer")

		return rowWriter.run(ctx, log)
	})

	return eg.Wait()
}

// a mapped record writer reads mapped records and writes them out
type mappedRecordWriter struct {
	Prefix     string
	Suffix     string
	LengthSize int

	w  *bufio.Writer
	ch <-chan mapper.MappedRecord
}

func (sr mappedRecordWriter) run(ctx context.Context, log *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			log.Warn("context was canceled")
			return context.Canceled
		case record, more := <-sr.ch:
			if !more {
				return nil
			}

			if err := sr.writeRow(ctx, log, record); err != nil {
				return err
			}
		}
	}
}

func (sr mappedRecordWriter) writeRow(ctx context.Context, log *zap.Logger, record mapper.MappedRecord) error {
	data, err := gojay.MarshalJSONObject(record)

	if err != nil {
		return err
	}

	if _, err := sr.w.WriteString(sr.Prefix); err != nil {
		return err
	}

	if sr.LengthSize > -1 {
		formattedJSONSize := strconv.FormatInt(int64(len(data)), 10) // A string of the size of the JSON object

		// If we've got a fixed number of bytes we want to write the JSON data in, we need to make sure it'll fit and then pad it out
		if sr.LengthSize > 0 {
			formattedJSONSizeLength := len(formattedJSONSize) // The length of the string of the size of the JSON object, so we can confirm it will fit without our fixed width

			toPad := sr.LengthSize - formattedJSONSizeLength

			if toPad < 0 {
				return err
			}

			// Write a bunch of zeros out
			for i := 0; i < toPad; i++ {
				if _, err := sr.w.WriteString("0"); err != nil {
					return err
				}
			}
		}

		// Whether or not we just padded it out, write the final JSON size prefix
		if _, err := sr.w.WriteString(formattedJSONSize); err != nil {
			return err
		}
	}

	// Write out the encoded data
	if _, err := sr.w.Write(data); err != nil {
		return err
	}

	// Write the suffix out - \n for json lines
	if _, err := sr.w.WriteString(sr.Suffix); err != nil {
		return err
	}

	if err := sr.w.Flush(); err != nil {
		return err
	}

	return nil
}
