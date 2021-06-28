package fmpxmlresult

import (
	"encoding/json"
)

// This is the internal FMPXMLResult data format, but with sane data types after conversion

type Product struct {
	Build   string `json:"build"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Database struct {
	DateFormat string `json:"dateFormat"`
	TimeFormat string `json:"timeFormat"`
	Layout     string `json:"layout"`
	Name       string `json:"name"`
	Records    int    `json:"records"`
}

type Field struct {
	EmptyOK   bool   `json:"emptyOK"`
	MaxRepeat int    `json:"maxRepeat"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type Metadata struct {
	Fields []Field `json:"fields,omitempty"`
}

type ResultSet struct {
	Found int   `json:"found"`
	Rows  []Row `json:"rows"`
}

type Col struct {
	Data []string `json:"data"`
}

// Row is the original FMPXMLRESULT row
type Row struct {
	ModID    string `json:"modID"`    // I'm not sure if this could not be an int, but as an identifier a string is A-OK
	RecordID string `json:"recordID"` // I'm not sure if this could not be an int, but as an identifier a string is A-OK
	Cols     []Col  `json:"cols"`
}

// Record is the normalized output for easy JSON work. The value may be a scalar or an array, based on the field repeat
type Record map[string]json.RawMessage

type FMPXMLResult struct {
	ErrorCode int        `json:"errorCode"`
	Product   *Product   `json:"product,omitempty"`
	Metadata  *Metadata  `json:"metadata,omitempty"`
	Database  *Database  `json:"database,omitempty"`
	ResultSet *ResultSet `json:"resultSet,omitempty"`

	// These will be set externally before PopulateRecords().
	// If they are set, the record ID and mod ID will be loaded into the corresponding field names
	RecordIDField   string `json:"-"` // The field to put the record ID in
	ModIDField      string `json:"-"` // The field to put the modification ID in
	SanitizeNumbers bool   `json:"-"` // If numbers should be reformatted, possibly with precision loss, or used as-is possibly (but unlikely) with compatibility issues

	Records []Record `json:"records,omitempty"`

	// These are used while populating the records
	dataEncoders         map[string]dataEncoder // The data encoders are how we change a DATE into a date, or a NUMBER into a number
	positionalColumnData []columnarData         // The positional column data includes both the column name and how to translate that particular field into a JSON object
}

// This will hold the information we need to reference by field order when iterating through the columns of a row
type columnarData struct {
	encoder fieldEncoder
	name    string
}
