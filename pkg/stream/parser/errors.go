package parser

import "fmt"

type wrappedError struct {
	error
}

func (err *wrappedError) Unwrap() error { return err.error }

type BooleanError struct {
	Original string
}

type IntegerDecodeError struct {
	wrappedError
	Original string
}

type NodeError struct {
	wrappedError
	Node string
}

func (err *BooleanError) Error() string {
	return fmt.Sprintf("Unable to decode '%s' as a boolean", err.Original)
}

func (err *IntegerDecodeError) Error() string {
	return fmt.Sprintf("Unable to decode '%s' as an integer: %v", err.Original, err.error)
}

func (err *NodeError) Error() string {
	return fmt.Sprintf("Error when handling %s: %v", err.Node, err.error)
}
