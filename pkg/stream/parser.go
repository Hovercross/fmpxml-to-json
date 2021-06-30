package stream

import (
	"context"
	"io"
	"log"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

type Parser struct {
	Reader io.Reader

	ErrorCodes chan<- int
	Products   chan<- fmpxmlresult.Product
	Fields     chan<- fmpxmlresult.Field
	Databases  chan<- fmpxmlresult.Database
	Rows       chan<- fmpxmlresult.NormalizedRow
}

func (p Parser) Parse(ctx context.Context) error {
	if p.ErrorCodes == nil {
		log.Println("Error codes is nil")
	}

	if p.Databases == nil {
		log.Println("Databases is nil")
	}

	if p.Fields == nil {
		log.Println("Fields is nil")
	}

	if p.Products == nil {
		log.Println("Products is nil")
	}

	if p.Rows == nil {
		log.Println("Products is ni")
	}

	parser := parser{
		ErrorCodes: p.ErrorCodes,
		Products:   p.Products,
		Fields:     p.Fields,
		Databases:  p.Databases,
		Rows:       p.Rows,
		ctx:        ctx,
	}

	return parser.Parse()
}
