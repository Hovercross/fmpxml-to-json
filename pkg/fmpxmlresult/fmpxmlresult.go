package fmpxmlresult

// This is the internal data format, but with sane data types

type Product struct {
	Build   string
	Name    string
	Version string
}

type Database struct {
	DateFormat string
	Layout     string
	Name       string
	Records    int
	TimeFormat string
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

type Row struct {
	ModID    string // I'm not sure if this could not be an int, but as an identifier a string is A-OK
	RecordID string // I'm not sure if this could not be an int, but as an identifier a string is A-OK
	Cols     []Col
}

type FMPXMLResult struct {
	ErrorCode int
	Product   Product
	Metadata  Metadata
	Database  Database
	ResultSet ResultSet
}
