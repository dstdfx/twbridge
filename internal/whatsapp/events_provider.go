package whatsapp

import (
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
	"go.uber.org/zap"
)

type EventsProvider struct {
	log         *zap.Logger
	whatsappConn *whatsapp.Conn
	outgoingEvents chan domain.Event
	startAt int64
}

type Opts struct {
	OutgoingEvents chan domain.Event
	WhatsappConn *whatsapp.Conn
}

func NewEventsProvider(log *zap.Logger, opts *Opts) *EventsProvider {
	return &EventsProvider{
		log:            log,
		startAt: time.Now().Unix(),
		outgoingEvents: opts.OutgoingEvents,
		whatsappConn: opts.WhatsappConn,
	}
}

func (wh *EventsProvider) HandleError(err error) {
	wh.log.Error("got error, trying to restore connection...", zap.Error(err))

	if err := wh.whatsappConn.Restore(); err != nil {
		wh.log.Error("failed to restore whatsapp connection", zap.Error(err))
	}
}

func (wh *EventsProvider) ShouldCallSynchronously() bool {
	return true
}

func (wh *EventsProvider) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.Timestamp < uint64(wh.startAt) || message.Info.FromMe {
		return
	}

	wh.log.Debug("got text message",
		zap.String("text", message.Text),
		zap.Bool("from_me", message.Info.FromMe),
		zap.Int("status", int(message.Info.Status)),
		zap.String("push_name", message.Info.PushName),
		zap.Uint64("timestamp", message.Info.Timestamp),
		zap.String("remote_jid", message.Info.RemoteJid),
		zap.String("sender_jid", message.Info.SenderJid))

	// Get the contact name from the contacts store
	var contactName string
	contacts := wh.whatsappConn.Store.Contacts
	contact, ok := contacts[message.Info.RemoteJid]
	if !ok {
		contactName = "<unknown>"
	} else {
		contactName = contact.Name
	}

	// TODO: handle outgoing events overflow
	select {
	case wh.outgoingEvents <- &domain.TextMessageEvent{
		WhatsappRemoteJid:  message.Info.RemoteJid,
		WhatsappSenderName: contactName,
		Text:               message.Text,
	}:
	default:
	}
}
