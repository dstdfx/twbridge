package whatsapp

import (
	"errors"
	"strings"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
	"go.uber.org/zap"
)

const (
	restoreAttempts = 3
	restoreInterval = time.Second
)

// EventsProvider represents whatsapp events provider.
type EventsProvider struct {
	log            *zap.Logger
	startAt        int64
	chatID         int64
	whatsappClient domain.WhatsappClient
	outgoingEvents chan domain.Event
}

// Opts represents options to create new instance of EventsProvider.
type Opts struct {
	// ChatID is identifier of telegram chat.
	ChatID int64

	// OutgoingEvents is a channel to send events to.
	OutgoingEvents chan domain.Event

	// WhatsappClient represents a client to work with whatsapp API.
	WhatsappClient domain.WhatsappClient
}

// NewEventsProvider creates new instance of EventsProvider.
func NewEventsProvider(log *zap.Logger, opts *Opts) *EventsProvider {
	return &EventsProvider{
		log:            log,
		chatID:         opts.ChatID,
		startAt:        time.Now().Unix(),
		outgoingEvents: opts.OutgoingEvents,
		whatsappClient: opts.WhatsappClient,
	}
}

//nolint
// HandleError method is called when error occurs.
func (wh *EventsProvider) HandleError(err error) {
	wh.log.Error("got error", zap.Error(err))

	// Ignore known errors that don't affect connection
	if strings.Contains(err.Error(), "error processing data: received invalid data") ||
		strings.Contains(err.Error(), "invalid string with tag 174") {

		return
	}

	switch err.(type) {
	case *whatsapp.ErrConnectionClosed, *whatsapp.ErrConnectionFailed:
		if !wh.restoreSession() {
			// TODO: send disconnect event
		}
	default:
		if errors.Is(err, whatsapp.ErrConnectionTimeout) {
			if !wh.restoreSession() {
				// TODO: send disconnect event
			}
		}
	}
}

func (wh *EventsProvider) restoreSession() (restored bool) {
	for i := 1; i <= restoreAttempts; i++ {
		wh.log.Debug("trying to restore whatsapp session...",
			zap.Int64("chat_id", wh.chatID),
			zap.Int("attempt", i))

		err := wh.whatsappClient.Restore()
		if err == nil {
			wh.startAt = time.Now().Unix()
			restored = true

			wh.log.Debug("session has been restored", zap.Int64("chat_id", wh.chatID))

			return
		}

		// TODO: fix to exponential backoff
		time.Sleep(restoreInterval)
	}

	wh.log.Error("failed to restore whatsapp session after several attempts",
		zap.Int64("chat_id", wh.chatID))

	return
}

// ShouldCallSynchronously method indicates how whatsapp events should be handled.
func (wh *EventsProvider) ShouldCallSynchronously() bool {
	return true
}

// HandleTextMessage method is called when new text message is received.
func (wh *EventsProvider) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.Timestamp < uint64(wh.startAt) || message.Info.FromMe {
		return
	}

	wh.log.Debug("got text message",
		zap.Bool("from_me", message.Info.FromMe),
		zap.Int("status", int(message.Info.Status)),
		zap.String("push_name", message.Info.PushName),
		zap.Uint64("timestamp", message.Info.Timestamp),
		zap.String("remote_jid", message.Info.RemoteJid),
		zap.String("sender_jid", message.Info.SenderJid))

	// Get the contact name from the contacts store
	var contactName string
	contacts := wh.whatsappClient.GetContacts()
	contact, ok := contacts[message.Info.RemoteJid]
	if !ok {
		contactName = "<unknown>"
	} else {
		contactName = contact.Name
	}

	wh.outgoingEvents <- &domain.TextMessageEvent{
		WhatsappRemoteJid:  message.Info.RemoteJid,
		WhatsappSenderName: contactName,
		Text:               message.Text,
		ChatID:             wh.chatID,
	}
}
