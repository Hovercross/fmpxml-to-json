package fmpxmlout

// The Product section of a Filemaker export
type Product struct {
	Build   string
	Name    string
	Version string
}

// The Database component
type Database struct {
	DateFormat string
	Layout     string
	Name       string
	Records    int
	TimeFormat string
}

type Metadata struct {
	Database  Database
	Product   Product
	ErrorCode int
	Fields    []*Field // This is a pointer because it will be cross-referenced
}

type Record struct {
	RecordID string
	ModID    string
	Data     map[*Field][]string
}

type Field struct {
	EmptyOK   bool
	MaxRepeat int
	Name      string
	Type      string
}

// A document representing an FMPXMLResult type input
type Document struct {
	Metadata Metadata
	Records  []Record
}
