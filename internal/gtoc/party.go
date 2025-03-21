package gtoc

import (
	"fmt"

	"github.com/invopop/gobl.cii/document"
	"github.com/invopop/gobl/catalogues/iso"
	"github.com/invopop/gobl/org"
)

// NewParty creates the SellerTradeParty part of a EN 16931 compliant invoice
func NewParty(party *org.Party) *document.Party {
	if party == nil {
		return nil
	}
	p := &document.Party{
		Name:                      party.Name,
		Contact:                   newContact(party),
		PostalTradeAddress:        newPostalTradeAddress(party.Addresses),
		URIUniversalCommunication: newEmail(party.Emails),
	}

	if party.TaxID != nil {
		// Assumes VAT ID being used instead of possible tax number
		p.SpecifiedTaxRegistration = []*document.SpecifiedTaxRegistration{
			{
				ID: &document.PartyID{
					Value:    party.TaxID.String(),
					SchemeID: "VA",
				},
			},
		}
	}

	if len(party.Identities) > 0 {
		for _, id := range party.Identities {
			if id.Ext.Has(iso.ExtKeySchemeID) {
				p.GlobalID = &document.PartyID{
					SchemeID: id.Ext[iso.ExtKeySchemeID].String(),
					Value:    id.Code.String(),
				}
			}
		}
	}

	if len(party.Inboxes) > 0 {
		p.URIUniversalCommunication = &document.URIUniversalCommunication{
			ID: &document.PartyID{
				Value:    party.Inboxes[0].Email,
				SchemeID: SchemeIDEmail,
			},
		}
	}

	return p
}

func newContact(p *org.Party) *document.Contact {
	if len(p.People) == 0 && len(p.Emails) == 0 && len(p.Telephones) == 0 {
		return nil
	}

	c := new(document.Contact)
	if len(p.People) > 0 {
		c.PersonName = contactName(p.People[0].Name)
	}

	if len(p.Telephones) > 0 {
		c.Phone = &document.PhoneNumber{
			CompleteNumber: p.Telephones[0].Number,
		}
	}

	if len(p.Emails) > 0 {
		c.Email = &document.Email{
			URIID: p.Emails[0].Address,
		}
	}

	return c
}

func contactName(p *org.Name) string {
	g := p.Given
	s := p.Surname

	if g == "" && s == "" {
		return ""
	}
	if g == "" {
		return s
	}
	if s == "" {
		return g
	}
	return fmt.Sprintf("%s %s", g, s)
}
