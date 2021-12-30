package whatsapp

import (
	"errors"

	"github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
)

var ErrUnsupportedMessageType = errors.New("got unsupported message type")

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

// Logout method invalidates the current whatsapp session.
func (c *Client) Logout() error {
	// We need to delete handlers in order to stop receiving updates from whatsapp.
	c.wc.RemoveHandlers()

	return c.wc.Logout()
}

// GetContacts method returns a list of whatsapp contacts.
func (c *Client) GetContacts() map[string]domain.WhatsappContact {
	if c.wc.Store == nil {
		return make(map[string]domain.WhatsappContact)
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
func (c *Client) Send(msg domain.WhatsappMessage) (err error) {
	var whatsappMessage interface{}
	switch msg.Type() {
	case domain.WhatsappTextMessageType:
		textMessage := msg.(*domain.WhatsappTextMessage)
		whatsappMessage = whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: textMessage.RemoteJid,
			},
			Text: textMessage.Text,
		}
	default:
		return ErrUnsupportedMessageType
	}

	_, err = c.wc.Send(whatsappMessage)

	return
}
