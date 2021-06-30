package stream

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/constants"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/paths"
)

// A parser emits normalizes and emits various records types as it reads the file. It does not do column -> field mapping
type parser struct {
	Reader io.Reader

	ErrorCodes chan<- int
	Products   chan<- fmpxmlresult.Product
	Fields     chan<- fmpxmlresult.Field
	Databases  chan<- fmpxmlresult.Database
	Rows       chan<- fmpxmlresult.NormalizedRow

	ctx          context.Context
	currentSpace paths.SpaceChain
	workingRow   fmpxmlresult.NormalizedRow
}

func (p *parser) Parse() error {
	decoder := xml.NewDecoder(p.Reader)

	for {
		// Check for cancelation before each token read
		select {
		case <-p.ctx.Done():
			return context.Canceled
		default:
		}

		token, err := decoder.Token()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		// Token may be one of StartElement, EndElement, Chardata, Commend, ProcInst, or Directive

		switch elem := token.(type) {
		case xml.StartElement:
			if err := p.handleStart(elem); err != nil {
				return err
			}
		case xml.EndElement:
			p.handleEnd(elem)
		case xml.CharData:
			if err := p.handleCharData(elem); err != nil {
				return err
			}
		}

	}
}

func (p *parser) handleStart(elem xml.StartElement) error {
	p.pushSpace(elem)

	if p.Products != nil && p.currentSpace.IsExact(paths.Product) {
		return p.handleProduct(elem)
	}

	if p.Databases != nil && p.currentSpace.IsExact(paths.Database) {
		return p.handleDatabase(elem)
	}

	if p.Fields != nil && p.currentSpace.IsExact(paths.Field) {
		return p.handleField(elem)
	}

	if p.Rows != nil && p.currentSpace.IsExact(paths.Row) {
		return p.handleRowStart(elem)
	}

	if p.Rows != nil && p.currentSpace.IsExact(paths.Col) {
		p.handleColStart(elem)
		return nil
	}

	log.Printf("Entering into %s", elem.Name.Local)
	return nil
}

func (p *parser) handleEnd(elem xml.EndElement) error {
	defer p.popSpace()

	log.Printf("Exiting out of %s", elem.Name.Local)

	if p.Rows != nil && p.currentSpace.IsExact(paths.Row) {
		return p.handleRowEnd()
	}

	return nil
}

func (p *parser) handleCharData(elem xml.CharData) error {
	log.Printf("Got data: %s", string(elem))

	if p.currentSpace.IsExact(paths.ErrorCode) {
		return p.handleErrorCode(elem)
	}

	if p.parseRows() && p.currentSpace.IsExact(paths.Data) {
		index := len(p.workingRow.Columns) - 1
		workingData := p.workingRow.Columns[index]
		workingData = append(workingData, string(elem))
		p.workingRow.Columns[index] = workingData
	}

	return nil
}

func (p *parser) handleErrorCode(elem xml.CharData) error {
	s := string(elem)
	val, err := strconv.Atoi(s)

	if err != nil {
		return fmt.Errorf("Unable to parse error code '%s' as integer: %v", s, err)
	}

	if p.ErrorCodes != nil {
		select {
		case <-p.ctx.Done():
			return context.Canceled
		case p.ErrorCodes <- val:
		}

	}

	return nil
}

func (p *parser) handleProduct(elem xml.StartElement) error {
	out := fmpxmlresult.Product{}

	attrMap := map[string]*string{
		constants.BUILD:   &out.Build,
		constants.NAME:    &out.Name,
		constants.VERSION: &out.Version,
	}

	for _, attr := range elem.Attr {
		if target, found := attrMap[attr.Name.Local]; found {
			*target = attr.Value
		}
	}
	select {
	case <-p.ctx.Done():
		return context.Canceled
	case p.Products <- out:
		return nil
	}
}

func (p *parser) handleDatabase(elem xml.StartElement) error {
	out := fmpxmlresult.Database{}

	attrMap := map[string]*string{
		constants.DATEFORMAT: &out.DateFormat,
		constants.LAYOUT:     &out.Layout,
		constants.NAME:       &out.Name,
		constants.TIMEFORMAT: &out.TimeFormat,
	}

	for _, attr := range elem.Attr {
		if target, found := attrMap[attr.Name.Local]; found {
			*target = attr.Value
		}

		if attr.Name.Local == constants.RECORDS {
			v, err := strconv.Atoi(attr.Value)
			if err != nil {
				return fmt.Errorf("Unable to parse records attribute '%s' as integer: %v", attr.Value, err)
			}

			out.Records = v
		}
	}

	select {
	case <-p.ctx.Done():
		return context.Canceled
	case p.Databases <- out:
		return nil
	}
}

func (p *parser) handleField(elem xml.StartElement) error {
	out := fmpxmlresult.Field{}

	for _, attr := range elem.Attr {
		if attr.Name.Local == constants.EMPTYOK {
			var err error
			if out.EmptyOK, err = yesNo(attr.Value); err != nil {
				return err
			}
		}

		if attr.Name.Local == constants.MAXREPEAT {
			var err error
			if out.MaxRepeat, err = strconv.Atoi(attr.Value); err != nil {
				return fmt.Errorf("Unable to parse '%s' as MAXREPEAT: %v", attr.Value, err)
			}
		}

		if attr.Name.Local == constants.NAME {
			out.Name = attr.Value
		}

		if attr.Name.Local == constants.TYPE {
			out.Type = attr.Value
		}
	}

	select {
	case <-p.ctx.Done():
		return context.Canceled
	case p.Fields <- out:
		return nil
	}
}

func (p *parser) handleRowStart(elem xml.StartElement) error {
	p.workingRow = fmpxmlresult.NormalizedRow{}

	for _, attr := range elem.Attr {
		if attr.Name.Local == constants.RECORDID {
			p.workingRow.RecordID = attr.Value
		}

		if attr.Name.Local == constants.MODID {
			p.workingRow.ModID = attr.Value
		}
	}

	return nil
}

func (p *parser) handleColStart(elem xml.StartElement) {
	// Create a new set of data elements for this row
	p.workingRow.Columns = append(p.workingRow.Columns, []string{})
}

// Once we've collected all the columns in a row, emit it
func (p *parser) handleRowEnd() error {
	select {
	case <-p.ctx.Done():
		return context.Canceled
	case p.Rows <- p.workingRow:
		return nil
	}

}

func (p *parser) parseRows() bool {
	return p.Rows != nil
}

func (p *parser) pushSpace(elem xml.StartElement) {
	p.currentSpace = append(p.currentSpace, elem.Name)
}

func (p *parser) popSpace() {
	p.currentSpace = p.currentSpace[0 : len(p.currentSpace)-1]
}

func yesNo(s string) (bool, error) {
	if s == "YES" {
		return true, nil
	}

	if s == "NO" {
		return false, nil
	}

	return false, fmt.Errorf("Unable to parse '%s' as YES/NO", s)
}
