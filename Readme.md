# FMPXMLResult to JSON

## Overview

This is a basic Go utility to convert a Filemaker FMPXMLResult type XML file into a JSON file. I was working with a client that has a very large Filemaker footprint that needed to sync their data with some external systems. Since the FMPXMLResult data format isn't the most friendly interchange format, this tool was written to take it and convert it to a much more "normal" JSON based format.

## Usage

There are five options to the utility controlling the input, output, and some formatting options.

- `-input`: The input file to read, defaulting to "-" for STDIN
- `-output`: The output file to write to, defaulting to "-" for STDOUT
- `-full`: Include a full original copy of the Filemaker data, including the metadata and result set
- `-recordID`: The field name to use for the Record ID
- `-modID`: The field name to use for the Modification ID
- `-reformatNumbers`: Reformat all numeric columns, in the event that there is a numeric format within the file that Go can parse but Javascript can not - only needed if there are issues ingesting numbers into the destination system

Basic usage example:

Generating `sample.json` from `sample.xml`, assuming the binary has been compiled to `fmpxml-to-json`:

`./fmpxml-to-json -input samples/sample.xml -output samples/sample.json -recordID recordID -modID modificationID`

In the vast majority of cases, applications wanting to ingest Filemaker data from the JSON form will look at the *records* field of the JSON output. This is in a "normal" format that most applications will be expecting - an array of dictionaries, where each dictionary is a row from the original file.

## Limitations

- Currently, the entire XML file needs to get read into memory before parsing and transforming the data. If very large files are being manipulated, the memory requirements of the binary will be proportionally large.

## Building

There is only one dependency currently, and it is only used for testing. To build the application, after installing Go, run `go build -o ./fmpxml-to-json cmd/fmpxml-to-json/*.go`

## References

- https://fmhelp.filemaker.com/help/12/fmp/en/html/import_export.17.33.html

