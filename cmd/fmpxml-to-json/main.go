package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/hovercross/fmpxml-to-json/pkg/mapper"
)

func main() {
	var inFileName, outFileName, recordIDField, modIDField string

	var stream bool
	var streamPrefix, streamSuffix string // These are only used for JSON Lines
	var streamLengthPrefixSize int

	flag.StringVar(&inFileName, "input", "-", "File to read from, or \"-\" for STDIN")
	flag.StringVar(&outFileName, "output", "-", "File to write to, or \"-\" for STDOUT")
	flag.StringVar(&recordIDField, "recordID", "", "Field name to write the record ID value to")
	flag.StringVar(&modIDField, "modID", "", "Field name to write the modification ID value to")
	flag.BoolVar(&stream, "json-stream", false, "Write a stream of JSON data instead of a single object")
	flag.StringVar(&streamPrefix, "json-stream-prefix", "", "Prefix to write before every entry in the JSON concatinated format")
	flag.StringVar(&streamSuffix, "json-stream-suffix", "\n", "Suffix to write before every entry in the JSON concatinated format")
	flag.IntVar(&streamLengthPrefixSize, "json-stream-prefix-size", -1, "Write the size of each JSON object before the JSON object itself. A value of 0 is an unlimited with, any other positive number indicates the fixed width for the JSON object size")

	flag.Parse()

	var reader io.ReadCloser

	if inFileName == "-" {
		reader = os.Stdin
	} else {
		var err error

		reader, err = os.Open(inFileName)

		if err != nil {
			log.Fatalf("Unable to open '%s' for reading: %s", inFileName, err)
		}
	}

	var writer io.WriteCloser

	if outFileName == "-" {
		writer = os.Stdout
	} else {
		var err error

		writer, err = os.Create(outFileName)

		if err != nil {
			log.Fatalf("Unable to open '%s' for writing: %s", outFileName, err)
		}
	}

	ctx := context.Background()

	// Default to the JSON format
	var f func() error = func() error {
		return mapper.WriteJSON(ctx, reader, writer, recordIDField, modIDField)
	}

	if stream {
		f = func() error {
			return mapper.WriteJSONLines(ctx, reader, writer, recordIDField, modIDField, streamPrefix, streamSuffix, streamLengthPrefixSize)
		}

	}

	if err := f(); err != nil {
		log.Fatal(err)
	}
}

func closeOrFatal(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Fatalf("Unable to close file: %v", err)
	}
}
