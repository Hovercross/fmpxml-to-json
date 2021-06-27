package fmpxmlresult

import (
	"encoding/json"
	"sync"
)

// This is the internal data format, but with sane data types

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
	EmptyOK   bool
	MaxRepeat int
	Name      string
	Type      string
}

type Metadata struct {
	Fields []Field
}

type ResultSet struct {
	Found int
	Rows  []Row
}

type Col struct {
	Data []string
}

// Row is the original FMPXMLRESULT row
type Row struct {
	ModID    string // I'm not sure if this could not be an int, but as an identifier a string is A-OK
	RecordID string // I'm not sure if this could not be an int, but as an identifier a string is A-OK
	Cols     []Col
}

// Record is the normalized output for easy JSON work. The value may be a scalar or an array, based on the field repeat
type Record map[string]json.RawMessage

type FMPXMLResult struct {
	ErrorCode int       `json:"errorCode"`
	Product   Product   `json:"product,omitempty"`
	Metadata  Metadata  `json:"metadata,omitempty"`
	Database  Database  `json:"database,omitempty"`
	ResultSet ResultSet `json:"resultSet,omitempty"`

	// These will be set externally before PopulateRecords().
	// If they are set, the record ID and mod ID will be loaded into the corresponding field names
	RecordIDField string `json:"-"`
	ModIDField    string `json:"-"`

	Records []Record `json:"records,omitempty"`

	dataEncoders         map[string]dataEncoder // The data encoders are how we change a DATE into a date, or a NUMBER into a number
	positionalColumnData []columnarData         // The positional column data includes both the column name and how to translate that particular field into a JSON object

	m sync.Mutex // Don't use this thing concurrently... but just in case it is, lock the internal state
}
