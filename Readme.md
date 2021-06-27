# FMPXMLResult to JSON

## Overview

This is a basic Go utility to convert a Filemaker FMPXMLResult type XML file into a JSON file. I was working with a client that has a very large Filemaker footprint that needed to sync their data with some external systems. Since the FMPXMLResult data format isn't the most friendly interchange format, this tool will take it and convert it to a much more "normal" JSON based format.

## Usage

There are five options to the utility controlling the input, output, and some formatting options.

- `-input`: The input file to read, defaulting to "-" for STDIN
- `-output`: The output file to write to, defaulting to "-" for STDOUT
- `-short`: Skip the mostly duplicate RecordSet output results
- `-recordID`: The field name to add the Record ID to
- `-modID`: The field name to add the Modification ID to

Basic usage example:

Generating `sample.json` from `sample.xml`, assuming the binary has been compiled to `fmpxml-to-json`:

`./fmpxml-to-json -input sample.xml -output sample.json`

## Building

There is only one dependency currently, and it is only used for testing. To build the application, after installing Go, run `go build -o ./fmpxml-to-json cmd/convert/main.go`

## References

- https://fmhelp.filemaker.com/help/12/fmp/en/html/import_export.17.33.html

