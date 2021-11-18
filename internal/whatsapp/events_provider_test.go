package whatsapp_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	whatsappsdk "github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/whatsapp"
	"github.com/dstdfx/twbridge/internal/whatsapp/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEventsProvider(t *testing.T) {
	t.Run("handle error", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		whatsappClientMock.On("Restore").Once().Return(nil)

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleError(errors.New("test error"))
		whatsappClientMock.AssertCalled(t, "Restore")
	})

	t.Run("handle text message", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		contacts := map[string]domain.WhatsappContact{
			"test-contact0-jid": {
				Jid:  "test-contact0-jid",
				Name: "test-contact0",
			},
			"test-contact1-jid": {
				Jid:  "test-contact1-jid",
				Name: "test-contact1",
			},
			"test-contact2-jid": {
				Jid:  "test-contact2-jid",
				Name: "test-contact2",
			},
		}
		testMessage := whatsappsdk.TextMessage{
			Info:        whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact1-jid",
				Timestamp: uint64(time.Now().Add(time.Minute).UnixNano()),
			},
			Text:        "test message",
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			gotEvent = <- outgoingEvents
			wg.Done()
		}()

		whatsappClientMock.On("GetContacts").Once().Return(contacts)

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleTextMessage(testMessage)
		whatsappClientMock.AssertCalled(t, "GetContacts")

		wg.Wait()

		// Check that event has been sent to outgoing channel
		assert.Equal(t, domain.TextMessageEventType, gotEvent.Type())

		// Check the event's content
		gotTextEvent := gotEvent.(*domain.TextMessageEvent)
		assert.Equal(t, testMessage.Text, gotTextEvent.Text)
		assert.Equal(t, testMessage.Info.RemoteJid, gotTextEvent.WhatsappRemoteJid)
		assert.Equal(t, contacts[testMessage.Info.RemoteJid].Name, gotTextEvent.WhatsappSenderName)
	})

	t.Run("handle text message, unknown user", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		contacts := map[string]domain.WhatsappContact{
			"test-contact0-jid": {
				Jid:  "test-contact0-jid",
				Name: "test-contact0",
			},
			"test-contact1-jid": {
				Jid:  "test-contact1-jid",
				Name: "test-contact1",
			},
			"test-contact2-jid": {
				Jid:  "test-contact2-jid",
				Name: "test-contact2",
			},
		}
		testMessage := whatsappsdk.TextMessage{
			Info:        whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact999-jid",
				Timestamp: uint64(time.Now().Add(time.Minute).UnixNano()),
			},
			Text:        "test message from unknown user",
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			gotEvent = <- outgoingEvents
			wg.Done()
		}()

		whatsappClientMock.On("GetContacts").Once().Return(contacts)

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleTextMessage(testMessage)
		whatsappClientMock.AssertCalled(t, "GetContacts")

		wg.Wait()

		// Check that event has been sent to outgoing channel
		assert.Equal(t, domain.TextMessageEventType, gotEvent.Type())

		// Check the event's content
		gotTextEvent := gotEvent.(*domain.TextMessageEvent)
		assert.Equal(t, testMessage.Text, gotTextEvent.Text)
		assert.Equal(t, testMessage.Info.RemoteJid, gotTextEvent.WhatsappRemoteJid)
		assert.Equal(t, "<unknown>", gotTextEvent.WhatsappSenderName)
	})

	t.Run("handle text message, ignore messages before handler start", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		testMessage := whatsappsdk.TextMessage{
			Info:        whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact1-jid",
				Timestamp: 1,
			},
			Text:        "test message",
		}

		whatsappClientMock.On("GetContacts")

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleTextMessage(testMessage)
		whatsappClientMock.AssertNotCalled(t, "GetContacts")
	})

	t.Run("handle text message, ignore self messages", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		testMessage := whatsappsdk.TextMessage{
			Info:        whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact1-jid",
				FromMe: true,
				Timestamp: uint64(time.Now().Add(time.Minute).UnixNano()),
			},
			Text:        "test message",
		}

		whatsappClientMock.On("GetContacts")

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleTextMessage(testMessage)
		whatsappClientMock.AssertNotCalled(t, "GetContacts")
	})
}
