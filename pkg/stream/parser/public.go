package parser

import (
	"context"
	"io"

	"go.uber.org/zap"
)

type Public struct {
	Rows         chan<- NormalizedRow
	ErrorCodes   chan<- ErrorCode
	Products     chan<- Product
	Fields       chan<- Field
	Databases    chan<- Database
	MetadataEnd  chan<- struct{}
	ResultSetEnd chan<- struct{}
}

func (p Public) getPrivate() *Parser {
	return &Parser{
		Rows:         p.Rows,
		ErrorCodes:   p.ErrorCodes,
		Products:     p.Products,
		Fields:       p.Fields,
		Databases:    p.Databases,
		MetadataEnd:  p.MetadataEnd,
		ResultSetEnd: p.ResultSetEnd,
	}
}

func (p Public) Parse(ctx context.Context, log *zap.Logger, r io.Reader) error {
	return p.getPrivate().Parse(ctx, log, r)
}
