package telegram

import (
	"context"
	"strings"

	"github.com/dstdfx/twbridge/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

// EventsProvider represents telegram events provider.
type EventsProvider struct {
	log               *zap.Logger
	eventsCh          chan domain.Event
	telegramUpdatesCh tgbotapi.UpdatesChannel
}

// Opts represents options to create new instance of EventsProvider.
type Opts struct {
	// TelegramUpdates is a channel to receive telegram updates from.
	TelegramUpdates tgbotapi.UpdatesChannel
}

// NewEventsProvider creates new instance of EventsProvider.
func NewEventsProvider(log *zap.Logger, opts *Opts) *EventsProvider {
	return &EventsProvider{
		log:               log,
		telegramUpdatesCh: opts.TelegramUpdates,
		eventsCh:          make(chan domain.Event, 1),
	}
}

// Run method starts the main goroutine of EventsProvider.
// The call is blocking.
func (ep *EventsProvider) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			ep.telegramUpdatesCh.Clear()
			close(ep.eventsCh)

			return nil
		case update := <-ep.telegramUpdatesCh:
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}

			// TODO: detect chat deletion

			switch update.Message.Text {
			case "/start":
				ep.eventsCh <- &domain.StartEvent{
					ChatID:   update.Message.Chat.ID,
					FromUser: update.Message.From.UserName,
				}
			case "/login":
				ep.eventsCh <- &domain.LoginEvent{
					ChatID:   update.Message.Chat.ID,
					FromUser: update.Message.From.UserName,
				}
			default:
				if update.Message.ReplyToMessage != nil {
					// Extract jid from the message that is replied to
					jidStart := strings.Index(update.Message.ReplyToMessage.Text, "jid:")
					jidEnd := strings.Index(update.Message.ReplyToMessage.Text, "]")

					// TODO: check borders properly
					if jidStart != -1 && jidEnd != -1 && jidStart < jidEnd {
						ep.eventsCh <- &domain.ReplyEvent{
							ChatID:    update.Message.Chat.ID,
							FromUser:  update.Message.From.UserName,
							Reply:     update.Message.Text,
							RemoteJid: update.Message.ReplyToMessage.Text[jidStart+5 : jidEnd],
						}
					}
				}
			}
		}
	}
}

// EventsStream method returns a stream of domain.Event.
func (ep *EventsProvider) EventsStream() <-chan domain.Event {
	return ep.eventsCh
}
