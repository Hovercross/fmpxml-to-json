package parser

type ErrorCode int

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

type NormalizedRow struct {
	ModID    string     `json:"modID"`
	RecordID string     `json:"recordID"`
	Columns  [][]string `json:"cols"` // 2d array, since each column has multiple data elements
}
