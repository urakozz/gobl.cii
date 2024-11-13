// Package gtoc contains the logic to convert a GOBL envelope into a CII document
package gtoc

import (
	"fmt"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.cii/document"
	"github.com/invopop/gobl/bill"
)

// Converter is the struct that contains the logic to convert a GOBL envelope into a CII document
type Converter struct {
	doc *document.Document
}

// Convert converts a GOBL envelope into a CIIdocument
func Convert(env *gobl.Envelope) (*document.Document, error) {
	c := new(Converter)
	c.doc = new(document.Document)
	inv, ok := env.Extract().(*bill.Invoice)
	if !ok {
		return nil, fmt.Errorf("invalid type %T", env.Document)
	}
	err := c.newDocument(inv)
	if err != nil {
		return nil, err
	}
	return c.doc, nil
}

func (c *Converter) newDocument(inv *bill.Invoice) error {

	c.doc = &document.Document{
		RSMNamespace: document.RSM,
		RAMNamespace: document.RAM,
		QDTNamespace: document.QDT,
		UDTNamespace: document.UDT,
		ExchangedContext: &document.ExchangedContext{
			GuidelineContext: &document.ExchangedContextParameter{ID: document.GuidelineContext},
		},
	}

	err := c.newHeader(inv)
	if err != nil {
		return err
	}

	err = c.newTransaction(inv)
	if err != nil {
		return err
	}

	return nil
}
