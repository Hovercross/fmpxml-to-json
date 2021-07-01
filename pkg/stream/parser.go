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

// A Parser emits normalizes and emits various records types as it reads the file. It does not do column -> field mapping
type Parser struct {
	Reader io.Reader

	ErrorCodes   chan<- int
	Products     chan<- fmpxmlresult.Product
	Fields       chan<- fmpxmlresult.Field
	Databases    chan<- fmpxmlresult.Database
	Rows         chan<- fmpxmlresult.NormalizedRow
	MetadataEnd  chan<- struct{} // Signaled whenever we end a metadata section
	ResultSetEnd chan<- struct{} // Signaled whenever we end a result set

	currentSpace paths.SpaceChain
	workingRow   fmpxmlresult.NormalizedRow
}

func (p *Parser) Parse(ctx context.Context) error {
	decoder := xml.NewDecoder(p.Reader)

	for {
		// Check for cancelation before each token read
		select {
		case <-ctx.Done():
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

	if p.Products != nil && p.currentSpace.IsExact(paths.Product) {
		return p.handleProduct(ctx, elem)
	}

	if p.Databases != nil && p.currentSpace.IsExact(paths.Database) {
		return p.handleDatabase(ctx, elem)
	}

	if p.Fields != nil && p.currentSpace.IsExact(paths.Field) {
		return p.handleField(ctx, elem)
	}

	if p.Rows != nil && p.currentSpace.IsExact(paths.Row) {
		return p.handleRowStart(ctx, elem)
	}

	if p.Rows != nil && p.currentSpace.IsExact(paths.Col) {
		p.handleColStart(ctx, elem)
		return nil
	}

	log.Printf("Entering into %s", elem.Name.Local)
	return nil
}

func (p *Parser) handleEnd(ctx context.Context, elem xml.EndElement) error {
	defer p.popSpace(ctx)

	log.Printf("Exiting out of %s", elem.Name.Local)

	if p.Rows != nil && p.currentSpace.IsExact(paths.Row) {
		return p.handleRowEnd(ctx)
	}

	if p.MetadataEnd != nil && p.currentSpace.IsExact(paths.Metadata) {
		p.handleMetadataEnd(ctx)
	}

	if p.ResultSetEnd != nil && p.currentSpace.IsExact(paths.ResultSet) {
		p.handleResultSetEnd(ctx)
	}

	return nil
}

func (p *Parser) handleCharData(ctx context.Context, elem xml.CharData) error {
	log.Printf("Got data: %s", string(elem))

	if p.currentSpace.IsExact(paths.ErrorCode) {
		return p.handleErrorCode(ctx, elem)
	}

	if p.Rows != nil && p.currentSpace.IsExact(paths.Data) {
		index := len(p.workingRow.Columns) - 1
		workingData := p.workingRow.Columns[index]
		workingData = append(workingData, string(elem))
		p.workingRow.Columns[index] = workingData
	}

	return nil
}

func (p *Parser) handleErrorCode(ctx context.Context, elem xml.CharData) error {
	s := string(elem)
	val, err := strconv.Atoi(s)

	if err != nil {
		return fmt.Errorf("Unable to parse error code '%s' as integer: %v", s, err)
	}

	if p.ErrorCodes != nil {
		select {
		case <-ctx.Done():
			return context.Canceled
		case p.ErrorCodes <- val:
		}

	}

	return nil
}

func (p *Parser) handleProduct(ctx context.Context, elem xml.StartElement) error {
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
	case <-ctx.Done():
		return context.Canceled
	case p.Products <- out:
		return nil
	}
}

func (p *Parser) handleDatabase(ctx context.Context, elem xml.StartElement) error {
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
	case <-ctx.Done():
		return context.Canceled
	case p.Databases <- out:
		return nil
	}
}

func (p *Parser) handleField(ctx context.Context, elem xml.StartElement) error {
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
	case <-ctx.Done():
		return context.Canceled
	case p.Fields <- out:
		return nil
	}
}

func (p *Parser) handleMetadataEnd(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case p.MetadataEnd <- struct{}{}:
		return nil
	}

}

func (p *Parser) handleResultSetEnd(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case p.ResultSetEnd <- struct{}{}:
		return nil
	}
}

func (p *Parser) handleRowStart(ctx context.Context, elem xml.StartElement) error {
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

func (p *Parser) handleColStart(ctx context.Context, elem xml.StartElement) {
	// Create a new set of data elements for this row
	p.workingRow.Columns = append(p.workingRow.Columns, []string{})
}

// Once we've collected all the columns in a row, emit it
func (p *Parser) handleRowEnd(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case p.Rows <- p.workingRow:
		return nil
	}
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
