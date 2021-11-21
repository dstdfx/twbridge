package whatsapp

import (
	"github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
)
// Client represents a whatsapp connection wrapper.
type Client struct {
	wc *whatsapp.Conn
}

// NewClient returns new instance of Client.
func NewClient(wc *whatsapp.Conn) *Client {
	return &Client{wc: wc}
}

// Restore method restores the current whatsapp session.
func (c *Client) Restore() error {
	return c.wc.Restore()
}

// GetContacts method returns a list of whatsapp contacts.
func (c *Client) GetContacts() map[string]domain.WhatsappContact {
	if c.wc.Store == nil {
		return make(map[string]domain.WhatsappContact, 0)
	}

	contacts := make(map[string]domain.WhatsappContact, len(c.wc.Store.Contacts))
	for jid, contact := range c.wc.Store.Contacts {
		contacts[jid] = domain.WhatsappContact{
			Jid:  contact.Jid,
			Name: contact.Name,
		}
	}

	return contacts
}

// Send method sends data via whatsapp client.
func (c *Client) Send(msg interface{}) (err error) {
	_, err = c.wc.Send(msg)

	return
}
