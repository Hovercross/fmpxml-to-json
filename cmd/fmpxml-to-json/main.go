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
	var full, sanitizeNumbers bool

	flag.StringVar(&inFileName, "input", "-", "File to read from, or \"-\" for STDIN")
	flag.StringVar(&outFileName, "output", "-", "File to write to, or \"-\" for STDOUT")
	flag.StringVar(&recordIDField, "recordID", "", "Field name to write the record ID value to")
	flag.StringVar(&modIDField, "modID", "", "Field name to write the modification ID value to")
	flag.BoolVar(&full, "full", false, "Keep all the original data")
	flag.BoolVar(&sanitizeNumbers, "reformatNumbers", false, "Reformat numbers for compatibility")

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

	if err := mapper.Map(context.Background(), reader, writer); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func closeOrFatal(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Fatalf("Unable to close file: %v", err)
	}
}
