package parser

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/constants"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/paths"
	"go.uber.org/zap"
)

// A Parser emits normalizes and emits various records types as it reads the file. It does not do column -> field mapping
type Parser struct {
	Rows         chan<- NormalizedRow
	ErrorCodes   chan<- ErrorCode
	Products     chan<- Product
	Fields       chan<- Field
	Databases    chan<- Database
	MetadataEnd  chan<- struct{}
	ResultSetEnd chan<- struct{}

	currentSpace paths.SpaceChain
	workingRow   NormalizedRow
}

func (p *Parser) Parse(ctx context.Context, log *zap.Logger, r io.Reader) error {
	log.Debug("starting XML decode")

	decoder := xml.NewDecoder(r)

	for {
		// Before we sit here decoding when nobody wants it, check for cancelation
		select {
		case <-ctx.Done():
			log.Warn("logging context canceled")
			return context.Canceled
		default:
		}

		token, err := decoder.Token()

		if err == io.EOF {
			log.Info("finished token parsing")
			return nil
		}

		if err != nil {
			log.Error("Error when getting token", zap.Error(err))
			return err
		}

		// Token may be one of StartElement, EndElement, Chardata, Command, ProcInst, or Directive

		log.Debug("beginning token switch")

		switch elem := token.(type) {
		case xml.StartElement:
			log := log.With(zap.String("token", elem.Name.Local))

			log.Debug("starting token")

			if err := p.handleStart(ctx, log, elem); err != nil {
				return err
			}
		case xml.EndElement:
			log := log.With(zap.String("token", elem.Name.Local))

			if err := p.handleEnd(ctx, log, elem); err != nil {
				return err
			}
		case xml.CharData:
			if err := p.handleCharData(ctx, log, elem); err != nil {
				return err
			}
		}
	}
}

func (p *Parser) handleStart(ctx context.Context, log *zap.Logger, elem xml.StartElement) error {
	p.pushSpace(ctx, elem)

	if p.currentSpace.IsExact(paths.Product) {
		return p.handleProduct(ctx, log, elem)
	}

	if p.currentSpace.IsExact(paths.Database) {
		return p.handleDatabase(ctx, log, elem)
	}

	if p.currentSpace.IsExact(paths.Field) {
		return p.handleField(ctx, log, elem)
	}

	if p.currentSpace.IsExact(paths.Row) {
		return p.handleRowStart(ctx, log, elem)
	}

	if p.currentSpace.IsExact(paths.Col) {
		return p.handleColStart(ctx, log, elem)
	}

	log.Debug("Token start was unhandled")

	return nil
}

func (p *Parser) handleEnd(ctx context.Context, log *zap.Logger, elem xml.EndElement) error {
	defer p.popSpace(ctx)

	log.Debug("ending token", zap.String("token", elem.Name.Local))

	if p.currentSpace.IsExact(paths.Row) {
		return p.handleRowEnd(ctx, log)
	}

	if p.currentSpace.IsExact(paths.Metadata) {
		return p.handleMetadataEnd(ctx, log)
	}

	if p.currentSpace.IsExact(paths.ResultSet) {
		return p.handleResultSetEnd(ctx, log)
	}

	log.Debug("token end not handled")

	return nil
}

func (p *Parser) handleCharData(ctx context.Context, log *zap.Logger, elem xml.CharData) error {
	if ce := log.Check(zap.DebugLevel, "handling char data"); ce != nil {
		ce.Write(zap.ByteString("charData", elem))
	}

	if p.currentSpace.IsExact(paths.ErrorCode) {
		return p.handleErrorCode(ctx, log, elem)
	}

	if p.Rows != nil && p.currentSpace.IsExact(paths.Data) {
		index := len(p.workingRow.Columns) - 1
		workingData := p.workingRow.Columns[index]
		workingData = append(workingData, string(elem))
		p.workingRow.Columns[index] = workingData
	}

	return nil
}

func (p *Parser) handleErrorCode(ctx context.Context, log *zap.Logger, elem xml.CharData) error {
	if ce := log.Check(zap.DebugLevel, "handling error code"); ce != nil {
		ce.Write(zap.ByteString("rawErrorCode", elem))
	}

	if p.ErrorCodes == nil {
		log.Debug("no error code handler")

		return nil
	}

	s := string(elem)
	val, err := strconv.Atoi(s)

	if err != nil {
		log.Error("could not translate error code into integer", zap.Error(err))
		return err
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled
	case p.ErrorCodes <- ErrorCode(val):
		return nil
	}
}

func (p *Parser) handleProduct(ctx context.Context, log *zap.Logger, elem xml.StartElement) error {
	log.Debug("starting product node handle")

	if p.Products == nil {
		log.Debug("product handler was nil")

		return nil
	}

	out := Product{}

	attrMap := map[string]*string{
		constants.BUILD:   &out.Build,
		constants.NAME:    &out.Name,
		constants.VERSION: &out.Version,
	}

	for _, attr := range elem.Attr {
		if ce := log.Check(zap.DebugLevel, "handling attribute"); ce != nil {
			ce.Write(zap.String("attr-key", attr.Name.Local), zap.String("attr-value", attr.Value))
		}

		if target, found := attrMap[attr.Name.Local]; found {
			*target = attr.Value
		}
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled

	case p.Products <- out:
		log.Debug("product emitted")
		return nil
	}
}

func (p *Parser) handleDatabase(ctx context.Context, log *zap.Logger, elem xml.StartElement) error {
	log.Debug("starting database node handle")

	if p.Databases == nil {
		log.Debug("database handler was nil")

		return nil
	}

	out := Database{}

	attrMap := map[string]*string{
		constants.DATEFORMAT: &out.DateFormat,
		constants.LAYOUT:     &out.Layout,
		constants.NAME:       &out.Name,
		constants.TIMEFORMAT: &out.TimeFormat,
	}

	for _, attr := range elem.Attr {
		if ce := log.Check(zap.DebugLevel, "handling attribute"); ce != nil {
			ce.Write(zap.String("attr-key", attr.Name.Local), zap.String("attr-value", attr.Value))
		}

		if target, found := attrMap[attr.Name.Local]; found {
			*target = attr.Value
		}

		if attr.Name.Local == constants.RECORDS {
			v, err := strconv.Atoi(attr.Value)
			if err != nil {
				log.Error("unable to parse records attribute as integer", zap.String("raw-value", attr.Value), zap.Error(err))

				return err
			}

			out.Records = v
		}
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled

	case p.Databases <- out:
		log.Debug("database was emitted")
		return nil
	}
}

func (p *Parser) handleField(ctx context.Context, log *zap.Logger, elem xml.StartElement) error {
	log.Debug("starting field node handle")

	if p.Fields == nil {
		log.Debug("field handler is nil")

		return nil
	}

	out := Field{}

	for _, attr := range elem.Attr {
		if attr.Name.Local == constants.EMPTYOK {
			var err error
			if out.EmptyOK, err = yesNo(attr.Value); err != nil {
				log.Error("unable to process EMPTYOK as boolean", zap.String("raw", attr.Value), zap.Error(err))

				return err
			}
		}

		if attr.Name.Local == constants.MAXREPEAT {
			var err error
			if out.MaxRepeat, err = strconv.Atoi(attr.Value); err != nil {
				log.Error("unable to process MAXREPEAT as integer", zap.String("raw", attr.Value), zap.Error(err))

				return err
			}
		}

		if attr.Name.Local == constants.NAME {
			out.Name = attr.Value
		}

		if attr.Name.Local == constants.TYPE {
			out.Type = attr.Value
		}
	}

	log.Debug("finishing parsing field, beginning emit")

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled

	case p.Fields <- out:
		log.Debug("field was emitted")
		return nil
	}
}

func (p *Parser) handleMetadataEnd(ctx context.Context, log *zap.Logger) error {
	log.Debug("handling metadata end")

	if p.MetadataEnd == nil {
		log.Debug("no metadata end handler")

		return nil
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled

	case p.MetadataEnd <- struct{}{}:
		log.Debug("metadata end was emitted")
		return nil
	}
}

func (p *Parser) handleResultSetEnd(ctx context.Context, log *zap.Logger) error {
	log.Debug("handling result set end")

	if p.ResultSetEnd == nil {
		log.Debug("no result set end handler")

		return nil
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled

	case p.ResultSetEnd <- struct{}{}:
		log.Debug("result set end was emitted")
		return nil
	}
}

func (p *Parser) handleRowStart(ctx context.Context, log *zap.Logger, elem xml.StartElement) error {
	log.Debug("handling row start")

	if p.Rows == nil {
		log.Debug("row handler is nil")

		return nil
	}

	p.workingRow = NormalizedRow{}

	for _, attr := range elem.Attr {
		log.Debug("handling attribute", zap.String("attr-name", attr.Name.Local), zap.String("attr-value", attr.Value))

		if attr.Name.Local == constants.RECORDID {
			p.workingRow.RecordID = attr.Value
		}

		if attr.Name.Local == constants.MODID {
			p.workingRow.ModID = attr.Value
		}
	}

	return nil
}

func (p *Parser) handleColStart(ctx context.Context, log *zap.Logger, elem xml.StartElement) error {
	log.Debug("handling col start")

	if p.Rows == nil {
		log.Debug("row handler is nil")

		return nil
	}

	// Create a new set of data elements for this row
	p.workingRow.Columns = append(p.workingRow.Columns, []string{})
	return nil
}

// Once we've collected all the columns in a row, emit it
func (p *Parser) handleRowEnd(ctx context.Context, log *zap.Logger) error {
	log.Debug("handling row end")

	if p.Rows == nil {
		log.Debug("row handler is nil")

		return nil
	}

	select {
	case <-ctx.Done():
		log.Warn("context was canceled")
		return context.Canceled

	case p.Rows <- p.workingRow:
		log.Debug("database was emitted")
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
