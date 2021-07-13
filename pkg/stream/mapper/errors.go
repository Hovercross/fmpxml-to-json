package mapper

import "fmt"

// An ExtraNodeError indicates we got more of a particular type of node than we expected - such as Metadata, Database, or Product
type ExtraNodeError struct {
	NodeName string
}

// A MissingNodeError indicates we didn't get an expected node
type MissingNodeError struct {
	NodeName string
}

// A FieldCountMismatch indicates we got a record with a column count that doesn't match the field count
type FieldCountMismatch struct {
	ParsedFields int
	RowFields    int
}

type MultipleScalarValuesError struct {
	DataCount int
}

// A RowError wraps another error and adds the row based data
type RowError struct {
	error
	RowIndex int
}

type ColError struct {
	error
	ColumnIndex int
}

type NumberDecodeError struct {
	Original string
}

type DateTimeParseError struct {
	error
	Layout string
	Input  string
}

type MissingDateTimeLayout struct{}

func (err *ExtraNodeError) Error() string {
	return fmt.Sprintf("Multiple %s nodes found", err.NodeName)
}

func (err *MissingNodeError) Error() string {
	return fmt.Sprintf("Missing node of type %s", err.NodeName)
}

func (err *FieldCountMismatch) Error() string {
	return "Mismatch between parsed fields and record columns"
}

func (err *RowError) Error() string {
	return fmt.Sprintf("Error when handling row %d: %v", err.RowIndex, err.error)
}

func (err *RowError) Unwrap() error {
	return err.error
}

func (err *ColError) Error() string {
	return fmt.Sprintf("Error when handling column %d: %v", err.ColumnIndex, err.error)
}

func (err *ColError) Unwrap() error {
	return err.error
}

func (err *MultipleScalarValuesError) Error() string {
	return fmt.Sprintf("Got %d data elements for a scalar field", err.DataCount)
}

func (err *NumberDecodeError) Error() string {
	return fmt.Sprintf("Error decoding '%s' as a number", err.Original)
}

func (err *DateTimeParseError) Error() string {
	return fmt.Sprintf("Error decoding '%s' as a datetime with layout '%s': %v", err.Input, err.Layout, err.error)
}

func (err *DateTimeParseError) Unwrap() error {
	return err.error
}

func (err *MissingDateTimeLayout) Error() string {
	return "Date/time layout was empty when parsing date/time field"
}
