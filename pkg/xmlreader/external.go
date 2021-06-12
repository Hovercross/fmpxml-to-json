package xmlreader

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
	Records    string
	TimeFormat string
}

type Metadata struct {
	Database  Database
	Product   Product
	ErrorCode int
}

type Record struct {
}

// A document representing an FMPXMLResult type input
type Document struct {
	Metadata Metadata
	Records  []Record
}
