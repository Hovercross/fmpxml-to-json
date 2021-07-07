package jsonStreamWrite

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/mapper"
)

var (
	ErrJSONTooLong = errors.New("When writing fixed-width length data, the JSON record would not fit inside said size")
)

func WriteJSONLines(ctx context.Context, r io.Reader, w io.Writer, recordIDField, modIDField, prefix, suffix string, lengthSize int) error {
	buf := bufio.NewWriter(w)

	write := func(ctx context.Context, record mapper.MappedRecord) error {
		data, err := gojay.MarshalJSONObject(record)

		if err != nil {
			return err
		}

		if _, err := buf.WriteString(prefix); err != nil {
			return err
		}

		if lengthSize > -1 {
			formattedJSONSize := strconv.FormatInt(int64(len(data)), 10) // A string of the size of the JSON object

			// If we've got a fixed number of bytes we want to write the JSON data in, we need to make sure it'll fit and then pad it out
			if lengthSize > 0 {
				formattedJSONSizeLength := len(formattedJSONSize) // The length of the string of the size of the JSON object, so we can confirm it will fit without our fixed width

				toPad := lengthSize - formattedJSONSizeLength

				if toPad < 0 {
					return ErrJSONTooLong
				}

				// Write a bunch of zeros out
				for i := 0; i < toPad; i++ {
					if _, err := buf.WriteString("0"); err != nil {
						return err
					}
				}
			}

			// Whether or not we just padded it out, write the final JSON size prefix
			if _, err := buf.WriteString(formattedJSONSize); err != nil {
				return nil
			}
		}

		// Write out the encoded data
		if _, err := buf.Write(data); err != nil {
			return err
		}

		// Write the suffix out - \n for json lines
		if _, err := buf.WriteString(suffix); err != nil {
			return err
		}

		if err := buf.Flush(); err != nil {
			return err
		}

		return nil
	}

	p := mapper.Mapper{
		RowIDField:          recordIDField,
		ModificationIDField: modIDField,
		RowHandler:          write,
	}

	// Run the mapper, which will in turn call our write function above
	_, err := p.Map(ctx, r)

	if err != nil {
		return err
	}

	return nil
}
