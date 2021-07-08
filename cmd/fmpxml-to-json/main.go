package main

import (
	"context"
	"flag"
	"io"
	nativeLog "log"
	"os"

	jsonWriter "github.com/hovercross/fmpxml-to-json/pkg/stream/writers/json"
	jsonStreamWriter "github.com/hovercross/fmpxml-to-json/pkg/stream/writers/json-stream"
	"go.uber.org/zap"
)

func main() {
	var inFileName, outFileName, recordIDField, modIDField string

	var stream bool
	var streamPrefix, streamSuffix string // These are only used for JSON Lines
	var streamLengthPrefixSize int

	var debug bool

	flag.StringVar(&inFileName, "input", "-", "File to read from, or \"-\" for STDIN")
	flag.StringVar(&outFileName, "output", "-", "File to write to, or \"-\" for STDOUT")
	flag.StringVar(&recordIDField, "recordID", "", "Field name to write the record ID value to")
	flag.StringVar(&modIDField, "modID", "", "Field name to write the modification ID value to")
	flag.BoolVar(&stream, "json-stream", false, "Write a stream of JSON data instead of a single object")
	flag.StringVar(&streamPrefix, "json-stream-prefix", "", "Prefix to write before every entry in the JSON concatinated format")
	flag.StringVar(&streamSuffix, "json-stream-suffix", "\n", "Suffix to write before every entry in the JSON concatinated format")
	flag.IntVar(&streamLengthPrefixSize, "json-stream-prefix-size", -1, "Write the size of each JSON object before the JSON object itself. A value of 0 is an unlimited with, any other positive number indicates the fixed width for the JSON object size")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")

	flag.Parse()

	log := getLog(debug)

	var reader io.ReadCloser

	if inFileName == "-" {
		reader = os.Stdin
	} else {
		var err error

		reader, err = os.Open(inFileName)

		if err != nil {
			log.Fatal("could not open file for reading", zap.String("filename", inFileName), zap.Error(err))
		}
	}

	var writer io.WriteCloser

	if outFileName == "-" {
		writer = os.Stdout
	} else {
		var err error

		writer, err = os.Create(outFileName)

		if err != nil {
			log.Fatal("could not open file for reading", zap.String("filename", outFileName), zap.Error(err))
		}
	}

	ctx := context.Background()

	// Default to the JSON format
	var f func() error = func() error {
		jw := jsonWriter.JSONWriter{
			RecordIDField: recordIDField,
			ModIDField:    modIDField,
		}

		return jw.Write(ctx, log, reader, writer)
	}

	if stream {
		f = func() error {
			sr := jsonStreamWriter.StreamWriter{
				RecordIDField: recordIDField,
				ModIDField:    modIDField,
				Prefix:        streamPrefix,
				Suffix:        streamSuffix,
				LengthSize:    streamLengthPrefixSize,
			}

			return sr.Write(ctx, log, reader, writer)
		}

	}

	if err := f(); err != nil {
		log.Fatal("unable to execute inner callable", zap.Error(err))
	}
}

func closeOrFatal(log *zap.Logger, f io.Closer) {
	if err := f.Close(); err != nil {
		log.Fatal("Unable to close file", zap.Error(err))
	}
}

func getLog(debug bool) *zap.Logger {
	if debug {
		log, err := zap.NewDevelopment()

		if err != nil {
			nativeLog.Fatalf("count not get Zap logger: %v", err)
		}

		return log
	}

	log, err := zap.NewProduction()
	if err != nil {
		nativeLog.Fatalf("count not get Zap logger: %v", err)
	}

	return log
}
