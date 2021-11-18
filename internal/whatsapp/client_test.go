package whatsapp_test

import (
	"testing"

	whatsappsdk "github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/whatsapp"
	"github.com/stretchr/testify/assert"
)

var _ domain.WhatsappClient = &whatsapp.Client{}

func TestClient_GetContacts(t *testing.T) {
	expectedContacts := map[string]domain.WhatsappContact{
		"1": {
			Jid:  "1",
			Name: "test1-name",
		},
		"2": {
			Jid:  "2",
			Name: "test2-name",
		},
		"3": {
			Jid:  "3",
			Name: "test3-name",
		},
	}
	testConn := &whatsappsdk.Conn{
		Store: &whatsappsdk.Store{
			Contacts: map[string]whatsappsdk.Contact{
				"1": {
					Jid:    "1",
					Notify: "test1-notify",
					Name:   "test1-name",
					Short:  "test1-short",
				},
				"2": {
					Jid:    "2",
					Notify: "test2-notify",
					Name:   "test2-name",
					Short:  "test2-short",
				},
				"3": {
					Jid:    "3",
					Notify: "test3-notify",
					Name:   "test3-name",
					Short:  "test3-short",
				},
			},
		},
	}
	testClient := whatsapp.NewClient(testConn)
	gotContacts := testClient.GetContacts()
	assert.Equal(t, expectedContacts, gotContacts)
}