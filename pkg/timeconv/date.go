package timeconv

// Conversions have to be ordered from most to least specific, so that yyyy doesn't become 0606
var dateConversions []conversion = []conversion{
	// Date parts
	{"yyyy", "2006"},
	{"yy", "06"},
	{"mm", "01"},
	{"MM", "01"},
	{"M", "1"},
	{"m", "1"},
	{"dd", "02"},
	{"d", "2"},
}

// ParseDateFormat will convert a Filemaker date format into a Go date format string
func ParseDateFormat(fmt string) string {
	return convert(fmt, dateConversions)
}
