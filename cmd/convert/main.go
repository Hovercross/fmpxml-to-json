package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/xmlreader"
)

func main() {
	var inFileName, outFileName string
	var short bool

	flag.StringVar(&inFileName, "input", "-", "File to read from, or \"-\" for STDIN")
	flag.StringVar(&outFileName, "output", "-", "File to write to, or \"-\" for STDOUT")
	flag.BoolVar(&short, "short", false, "Remove the original record set")
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

	parsed, parseErr := xmlreader.ReadXML(reader)

	closeOrFatal(reader)

	if parseErr != nil {
		log.Fatalf("Unable to read input XML: %v", parseErr)
	}

	if err := parsed.PopulateRecords(); err != nil {
		log.Fatalf("Unable to convert record format: %v", err)
	}

	if short {
		parsed.ResultSet = fmpxmlresult.ResultSet{}
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

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(parsed); err != nil {
		closeOrFatal(writer)

		log.Fatalf("Could not write JSON data: %v", err)
	}

	closeOrFatal(writer)
}

func closeOrFatal(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Fatalf("Unable to close file: %v", err)
	}
}
