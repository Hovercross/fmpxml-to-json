package encoder

// An Encoder translates from Filemaker formats to JSON formats
type Encoder struct {
	parseNumbers    bool
	dateLayout      string
	timeLayout      string
	timestampLayout string
}

type FieldEncoder struct {
	array bool
}
