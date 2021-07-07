package parser

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

// A Parser emits normalizes and emits various records types as it reads the file. It does not do column -> field mapping
type Parser struct {
	Reader io.Reader

	ErrorCodeHandler    func(context.Context, fmpxmlresult.ErrorCode) error
	ProductHandler      func(context.Context, fmpxmlresult.Product) error
	FieldHandler        func(context.Context, fmpxmlresult.Field) error
	DatabaseHandler     func(context.Context, fmpxmlresult.Database) error
	RowHandler          func(context.Context, fmpxmlresult.NormalizedRow) error
	MetadataEndHandler  func(context.Context) error
	ResultSetEndHandler func(context.Context) error

	currentSpace paths.SpaceChain
	workingRow   fmpxmlresult.NormalizedRow
}

func (p *Parser) Parse(ctx context.Context) error {
	decoder := xml.NewDecoder(p.Reader)

	for {
		token, err := decoder.Token()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		// Token may be one of StartElement, EndElement, Chardata, Command, ProcInst, or Directive

		switch elem := token.(type) {
		case xml.StartElement:
			if err := p.handleStart(ctx, elem); err != nil {
				return err
			}
		case xml.EndElement:
			p.handleEnd(ctx, elem)
		case xml.CharData:
			if err := p.handleCharData(ctx, elem); err != nil {
				return err
			}
		}

	}
}

func (p *Parser) handleStart(ctx context.Context, elem xml.StartElement) error {
	p.pushSpace(ctx, elem)

	log.Printf("Entering into %s", elem.Name.Local)

	if p.currentSpace.IsExact(paths.Product) {
		return p.handleProduct(ctx, elem)
	}

	if p.currentSpace.IsExact(paths.Database) {
		return p.handleDatabase(ctx, elem)
	}

	if p.currentSpace.IsExact(paths.Field) {
		return p.handleField(ctx, elem)
	}

	if p.currentSpace.IsExact(paths.Row) {
		return p.handleRowStart(ctx, elem)
	}

	if p.currentSpace.IsExact(paths.Col) {
		return p.handleColStart(ctx, elem)
	}

	log.Printf("Element was unhandled")

	return nil
}

func (p *Parser) handleEnd(ctx context.Context, elem xml.EndElement) error {
	defer p.popSpace(ctx)

	log.Printf("Exiting out of %s", elem.Name.Local)

	if p.currentSpace.IsExact(paths.Row) {
		return p.handleRowEnd(ctx)
	}

	if p.currentSpace.IsExact(paths.Metadata) {
		return p.handleMetadataEnd(ctx)
	}

	if p.currentSpace.IsExact(paths.ResultSet) {
		return p.handleResultSetEnd(ctx)
	}

	log.Printf("Element was unhandled")

	return nil
}

func (p *Parser) handleCharData(ctx context.Context, elem xml.CharData) error {
	log.Printf("Got data: %s", string(elem))

	if p.currentSpace.IsExact(paths.ErrorCode) {
		return p.handleErrorCode(ctx, elem)
	}

	if p.RowHandler != nil && p.currentSpace.IsExact(paths.Data) {
		index := len(p.workingRow.Columns) - 1
		workingData := p.workingRow.Columns[index]
		workingData = append(workingData, string(elem))
		p.workingRow.Columns[index] = workingData
	}

	return nil
}

func (p *Parser) handleErrorCode(ctx context.Context, elem xml.CharData) error {
	if p.ErrorCodeHandler == nil {
		return nil
	}

	s := string(elem)
	val, err := strconv.Atoi(s)

	if err != nil {
		return fmt.Errorf("Unable to parse error code '%s' as integer: %v", s, err)
	}

	return p.ErrorCodeHandler(ctx, fmpxmlresult.ErrorCode(val))
}

func (p *Parser) handleProduct(ctx context.Context, elem xml.StartElement) error {
	if p.ProductHandler == nil {
		return nil
	}

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

	return p.ProductHandler(ctx, out)
}

func (p *Parser) handleDatabase(ctx context.Context, elem xml.StartElement) error {
	if p.DatabaseHandler == nil {
		return nil
	}

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

	return p.DatabaseHandler(ctx, out)
}

func (p *Parser) handleField(ctx context.Context, elem xml.StartElement) error {
	if p.FieldHandler == nil {
		return nil
	}

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

	return p.FieldHandler(ctx, out)
}

func (p *Parser) handleMetadataEnd(ctx context.Context) error {
	if p.MetadataEndHandler == nil {
		return nil
	}

	return p.MetadataEndHandler(ctx)
}

func (p *Parser) handleResultSetEnd(ctx context.Context) error {
	if p.ResultSetEndHandler == nil {
		return nil
	}

	return p.ResultSetEndHandler(ctx)
}

func (p *Parser) handleRowStart(ctx context.Context, elem xml.StartElement) error {
	if p.RowHandler == nil {
		return nil
	}

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

func (p *Parser) handleColStart(ctx context.Context, elem xml.StartElement) error {
	if p.RowHandler == nil {
		return nil
	}

	// Create a new set of data elements for this row
	p.workingRow.Columns = append(p.workingRow.Columns, []string{})
	return nil
}

// Once we've collected all the columns in a row, emit it
func (p *Parser) handleRowEnd(ctx context.Context) error {
	if p.RowHandler == nil {
		return nil
	}

	return p.RowHandler(ctx, p.workingRow)
}

func (p *Parser) pushSpace(ctx context.Context, elem xml.StartElement) {
	p.currentSpace = append(p.currentSpace, elem.Name)
}

func (p *Parser) popSpace(ctx context.Context) {
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
