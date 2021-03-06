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
	testChatID := int64(123)

	t.Run("handle error, ignored", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleError(errors.New("error processing data: received invalid data")) //nolint
		eventsProvider.HandleError(errors.New("invalid string with tag 174"))                  //nolint
		whatsappClientMock.AssertNotCalled(t, "Restore")
	})

	t.Run("handle error, session restored", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		whatsappClientMock.On("Restore").Return(nil)
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleError(whatsappsdk.ErrConnectionTimeout)
		whatsappClientMock.AssertCalled(t, "Restore")
	})

	t.Run("handle error, failed to restore session", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		whatsappClientMock.On("Restore").Return(errors.New("failed to restore session")) // nolint
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			ChatID:         testChatID,
			OutgoingEvents: outgoingEvents,
			WhatsappClient: whatsappClientMock,
		})

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			gotEvent = <-outgoingEvents
			wg.Done()
		}()

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleError(whatsappsdk.ErrConnectionTimeout)
		whatsappClientMock.AssertCalled(t, "Restore")

		wg.Wait()

		// Check that event has been sent to outgoing channel
		assert.Equal(t, domain.DisconnectEventType, gotEvent.Type())

		// Check the event's content
		gotDisconnectEvent := gotEvent.(*domain.DisconnectEvent)
		assert.Equal(t, testChatID, gotDisconnectEvent.ChatID)
	})

	t.Run("handle text message", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			ChatID:         testChatID,
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
			Info: whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact1-jid",
				Timestamp: uint64(time.Now().Add(time.Minute).UnixNano()),
			},
			Text: "test message",
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			gotEvent = <-outgoingEvents
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
		assert.Equal(t, testChatID, gotTextEvent.ChatID)
		assert.Equal(t, testMessage.Text, gotTextEvent.Text)
		assert.Equal(t, testMessage.Info.RemoteJid, gotTextEvent.WhatsappRemoteJid)
		assert.Equal(t, contacts[testMessage.Info.RemoteJid].Name, gotTextEvent.WhatsappSenderName)
	})

	t.Run("handle text message, unknown user", func(t *testing.T) {
		// Init test events provider
		outgoingEvents := make(chan domain.Event, 1)
		whatsappClientMock := &mocks.WhatsappClient{}
		eventsProvider := whatsapp.NewEventsProvider(zap.NewNop(), &whatsapp.Opts{
			ChatID:         testChatID,
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
			Info: whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact999-jid",
				Timestamp: uint64(time.Now().Add(time.Minute).UnixNano()),
			},
			Text: "test message from unknown user",
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			gotEvent = <-outgoingEvents
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
		assert.Equal(t, testChatID, gotTextEvent.ChatID)
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
			Info: whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact1-jid",
				Timestamp: 1,
			},
			Text: "test message",
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
			Info: whatsappsdk.MessageInfo{
				Id:        "1",
				RemoteJid: "test-contact1-jid",
				FromMe:    true,
				Timestamp: uint64(time.Now().Add(time.Minute).UnixNano()),
			},
			Text: "test message",
		}

		whatsappClientMock.On("GetContacts")

		// Call method in order to emulate whatsapp event
		eventsProvider.HandleTextMessage(testMessage)
		whatsappClientMock.AssertNotCalled(t, "GetContacts")
	})
}
