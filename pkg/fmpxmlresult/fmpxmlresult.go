package fmpxmlresult

import "encoding/json"

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

// Record is the normalized output for easy JSON work
type Record map[string]json.RawMessage

type FMPXMLResult struct {
	ErrorCode int       `json:"errorCode"`
	Product   Product   `json:"product"`
	Metadata  Metadata  `json:"metadata"`
	Database  Database  `json:"database"`
	ResultSet ResultSet `json:"resultSet"`

	Records []Record `json:"records,omitempty"`
}
