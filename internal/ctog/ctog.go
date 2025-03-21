// Package ctog contains the logic to convert a CII document into a GOBL envelope
package ctog

import (
	"github.com/nbio/xml"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.cii/document"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/catalogues/untdid"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

// Converter is a struct that contains the necessary elements to convert between GOBL and CII
type Converter struct {
	// CtoG Output
	inv *bill.Invoice
	// CtoG Input
	doc *document.Invoice
}

// Convert converts a CII document into a GOBL envelope
func Convert(xmlData []byte) (*gobl.Envelope, error) {
	c := new(Converter)
	c.inv = new(bill.Invoice)
	c.doc = new(document.Invoice)
	if err := xml.Unmarshal(xmlData, &c.doc); err != nil {
		return nil, err
	}

	err := c.NewInvoice(c.doc)
	if err != nil {
		return nil, err
	}

	env, err := gobl.Envelop(c.inv)
	if err != nil {
		return nil, err
	}
	return env, nil
}

// NewInvoice creates a new GOBL invoice from a CII document
func (c *Converter) NewInvoice(doc *document.Invoice) error {

	c.inv = &bill.Invoice{
		Code:     cbc.Code(doc.ExchangedDocument.ID),
		Type:     TypeCodeParse(doc.ExchangedDocument.TypeCode),
		Currency: currency.Code(doc.Transaction.Settlement.Currency),
		Supplier: c.getParty(doc.Transaction.Agreement.Seller),
		Customer: c.getParty(doc.Transaction.Agreement.Buyer),
		Tax: &bill.Tax{
			Ext: tax.Extensions{
				untdid.ExtKeyDocumentType: cbc.Code(doc.ExchangedDocument.TypeCode),
			},
		},
	}

	issueDate, err := ParseDate(doc.ExchangedDocument.IssueDate.DateFormat.Value)
	if err != nil {
		return err
	}
	c.inv.IssueDate = issueDate

	err = c.prepareLines(doc.Transaction)
	if err != nil {
		return err
	}

	// Payment comprised of terms, means and payee. Check tehre is relevant info in at least one of them to create a payment
	ahts := doc.Transaction.Settlement
	if ahts.HasPayment() {
		err = c.preparePayment(ahts)
		if err != nil {
			return err
		}
	}

	if len(doc.ExchangedDocument.IncludedNote) > 0 {
		c.inv.Notes = make([]*org.Note, 0, len(doc.ExchangedDocument.IncludedNote))
		for _, note := range doc.ExchangedDocument.IncludedNote {
			n := &org.Note{
				Text: note.Content,
			}
			if note.SubjectCode != "" {
				n.Code = cbc.Code(note.SubjectCode)
			}
			c.inv.Notes = append(c.inv.Notes, n)
		}
	}

	err = c.prepareOrdering(doc)
	if err != nil {
		return err
	}

	err = c.prepareDelivery(doc.Transaction.Delivery)
	if err != nil {
		return err
	}

	if len(ahts.ReferencedDocument) > 0 {
		c.inv.Preceding = make([]*org.DocumentRef, 0, len(ahts.ReferencedDocument))
		for _, ref := range ahts.ReferencedDocument {
			docRef := &org.DocumentRef{
				Code: cbc.Code(ref.IssuerAssignedID),
			}
			if ref.IssueDate != nil && ref.IssueDate.DateFormat != nil {
				refDate, err := ParseDate(ref.IssueDate.DateFormat.Value)
				if err != nil {
					return err
				}
				docRef.IssueDate = &refDate
			}
			c.inv.Preceding = append(c.inv.Preceding, docRef)
		}
	}

	if doc.Transaction.Agreement.TaxRepresentative != nil {
		// Move the original seller to the ordering.seller party
		if c.inv.Ordering == nil {
			c.inv.Ordering = &bill.Ordering{}
		}
		c.inv.Ordering.Seller = c.inv.Supplier

		// Overwrite the seller field with the tax representative
		c.inv.Supplier = c.getParty(doc.Transaction.Agreement.TaxRepresentative)
	}

	if len(ahts.AllowanceCharges) > 0 {
		err = c.prepareChargesAndDiscounts(ahts)
		if err != nil {
			return err
		}
	}
	return nil
}
