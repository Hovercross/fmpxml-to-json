package timeconv

var timeConversions []conversion = []conversion{
	{"hh", "03"},
	{"h", "3"},
	{"kk", "15"},
	{"k", "15"},
	{"mm", "04"},
	{"ss", "05"},
	{"a", "PM"},
}

// ParseTimeFormat will convert a Filemaker time format into a Go time format string
func ParseTimeFormat(fmt string) string {
	return convert(fmt, timeConversions)
}
