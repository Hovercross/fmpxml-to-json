package fmpxmlresult

import (
	"encoding/json"
	"fmt"

	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
)

// A fieldEncoder will take an entire column, which may or may not have repeating <DATA> elements,
// and put it into a single JSON value or an array, based on the MAXREPEAT values
type fieldEncoder func([]string) (json.RawMessage, error)

// PopulateRecords will create all the easy to read record data
func (fmp *FMPXMLResult) PopulateRecords() error {
	fmp.m.Lock()
	defer fmp.m.Unlock()

	fmp.populateFieldEncoders()

	// Empty out our record destination, and allocate it in a single go
	fmp.Records = make([]Record, len(fmp.ResultSet.Rows))

	for i, row := range fmp.ResultSet.Rows {
		record := Record{}

		if fmp.RecordIDField != "" {
			recordID, _ := json.Marshal(row.RecordID) // You can never fail to marshal a string
			record[fmp.RecordIDField] = recordID
		}

		if fmp.ModIDField != "" {
			modID, _ := json.Marshal(row.ModID) // You can never fail to marshal a string
			record[fmp.ModIDField] = modID
		}

		if len(row.Cols) != len(fmp.positionalColumnData) {
			return fmt.Errorf("Row %d column count mismatch: have %d, expect %d", i, len(row.Cols), len(fmp.positionalColumnData))
		}

		for j, col := range row.Cols {
			positionalItem := fmp.positionalColumnData[j]

			encoder := positionalItem.encoder
			name := positionalItem.name

			encoded, err := encoder(col.Data)

			if err != nil {
				return fmt.Errorf("Unable to encode row %d column %d: %v", i, j, err)
			}

			record[name] = encoded
		}

		fmp.Records[i] = record
	}

	return nil
}

func (fmp *FMPXMLResult) populateFieldEncoders() {
	fmp.populateDataEncoders()

	fmp.positionalColumnData = make([]columnarData, len(fmp.Metadata.Fields))

	// Load each of the encoders
	for i, field := range fmp.Metadata.Fields {
		fmp.positionalColumnData[i] = columnarData{
			fmp.getEncoder(field),
			field.Name,
		}
	}
}

func (fmp *FMPXMLResult) populateDataEncoders() {
	dateFormat := timeconv.ParseDateFormat(fmp.Database.DateFormat)
	timeFormat := timeconv.ParseTimeFormat(fmp.Database.TimeFormat)
	timestampFormat := dateFormat + " " + timeFormat // No idea if this is right, don't have an example handy

	// The specific datum normalizers we will be using for this file
	fmp.dataEncoders = map[string]dataEncoder{
		"DATE":      getTimeEncoder(dateFormat, "2006-01-02"),
		"TIME":      getTimeEncoder(timeFormat, "15:04:05"),
		"TIMESTAMP": getTimeEncoder(timestampFormat, "2006-01-02T15:04:05"),
		"NUMBER":    encodeNumber,
	}
}

// Get the encoder for a given field
func (fmp *FMPXMLResult) getEncoder(f Field) fieldEncoder {
	var dn dataEncoder = encodeString // Just a default

	// Override the string encoder, if we have a more appropriate encoder
	if v, found := fmp.dataEncoders[f.Type]; found {
		dn = v
	}

	if f.MaxRepeat == 1 {
		return getScalarEncoder(dn)
	}

	return getArrayEncoder(dn)
}
