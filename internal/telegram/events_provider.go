package telegram

import (
	"context"

	"github.com/dstdfx/twbridge/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

// EventsProvider represents telegram events provider.
type EventsProvider struct {
	log         *zap.Logger
	telegramAPI *tgbotapi.BotAPI
	eventsCh    chan domain.Event
}

// Opts represents options to create new instance of EventsProvider.
type Opts struct {
	TelegramAPI *tgbotapi.BotAPI
}

// NewEventsProvider creates new instance of EventsProvider.
func NewEventsProvider(log *zap.Logger, opts *Opts) *EventsProvider {
	return &EventsProvider{
		log:         log,
		telegramAPI: opts.TelegramAPI,
		eventsCh:    make(chan domain.Event),
	}
}

// Run method starts the main goroutine of EventsProvider.
// The call is blocking.
func (ep *EventsProvider) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := ep.telegramAPI.GetUpdatesChan(u)
	if err != nil {
		ep.log.Error("failed to get updates chan", zap.Error(err))

		return err
	}

	for {
		select {
		case <-ctx.Done():
			ep.telegramAPI.StopReceivingUpdates()
			close(ep.eventsCh)

			return nil
		case update := <-updates:
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
				// TODO: handle reply events
			}
		}
	}
}

// EventsStream method returns a stream of domain.Event.
func (ep *EventsProvider) EventsStream() <-chan domain.Event {
	return ep.eventsCh
}

// Type method returns a type of the provider.
func (ep *EventsProvider) Type() domain.EventProviderType {
	return domain.TelegramEventProvider
}
