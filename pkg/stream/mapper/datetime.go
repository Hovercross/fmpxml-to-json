package mapper

import "time"

func reformatDT(inputVal, inputLayout, outputLayout string) (string, error) {
	if inputLayout == "" {
		return "", ErrNoTimeParseLayout
	}
	dt, err := time.Parse(inputLayout, inputVal)

	if err != nil {
		return "", err
	}

	return dt.Format(outputLayout), nil
}

func (m *mapper) reformatDate(s string) (string, error) {
	return reformatDT(s, m.dateLayout, "2006-01-02")
}

func (m *mapper) reformatTime(s string) (string, error) {
	return reformatDT(s, m.timeLayout, "15:04:05")
}

func (m *mapper) reformatTimestamp(s string) (string, error) {
	return reformatDT(s, m.timestampLayout, "2006-01-02 15:04:05")
}
