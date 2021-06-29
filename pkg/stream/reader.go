package stream

import (
	"bufio"
	"io"

	"github.com/francoispqt/gojay"
	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
	xmlparser "github.com/tamerh/xml-stream-parser"
)

func Parse(r io.Reader, out chan<- encodeList) {
	buf := bufio.NewReaderSize(r, 65536)
	xml := xmlparser.NewXMLParser(buf, "DATABASE", "FIELD", "ROW").EnableXpath()

	p := &parser{
		xml: xml,
	}

	p.parse()
}

type parser struct {
	xml *xmlparser.XMLParser

	fields encodeList

	dateTranslator      func(string) (string, error)
	timeTranslator      func(string) (string, error)
	timestampTranslator func(string) (string, error)
}

func (p *parser) parse() error {
	for xml := range p.xml.Stream() {
		if xml.Name == "FIELD" {
			p.handleField(xml)
		}

		if xml.Name == "DATABASE" {
			p.handleDatabase(xml)
		}

		if xml.Name == "ROW" {
			p.handleRow(xml)
		}

	}

	return nil
}

func (p *parser) handleField(xml *xmlparser.XMLElement) {

}

func (p *parser) getEncoder(xml *xmlparser.XMLElement) encode {
	// name := xml.Attrs["NAME"]
	// maxRepeat := xml.Attrs["MAXREPEAT"]
	// fieldType := xml.Attrs["TYPE"]

	// switch fieldType {
	// case "DATE":

	// case "TIME":
	// case "TIMESTAMP":
	// case "NUMBER":
	// default:
	// 	return func(enc *gojay.Encoder) { enc.AddStringKey(name) }
	// }

	return nil
}

func (p *parser) handleDatabase(xml *xmlparser.XMLElement) {
	dateFormat := xml.Attrs["DATEFORMAT"]
	timeFormat := xml.Attrs["TIMEFORMAT"]

	referenceDate := timeconv.ParseDateFormat(dateFormat)
	referenceTime := timeconv.ParseTimeFormat(timeFormat)
	referenceTimestamp := referenceDate + " " + referenceTime // Maybe?

	p.dateTranslator = timeconv.MakeTranslationFunc(referenceDate, "2006-01-02")
	p.timeTranslator = timeconv.MakeTranslationFunc(referenceTime, "15:04:05")
	p.timestampTranslator = timeconv.MakeTranslationFunc(referenceTimestamp, "2006-01-02 15:04:05")
}

func (p *parser) handleRow(xml *xmlparser.XMLElement) {

}

func makeTimeTranslation(layout, format string) (*gojay.EmbeddedJSON, error) {
	return nil, nil
}
