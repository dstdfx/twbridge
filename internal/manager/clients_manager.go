package manager

import (
	"context"

	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

// Manager handles incoming events from telegram provider and manages new
// and existing clients of the bot.
type Manager struct {
	log            *zap.Logger
	incomingEvents <-chan domain.Event
	telegramAPI    *tgbotapi.BotAPI
	eventHandlers  map[int64]chan<- domain.Event
}

// Opts represents options to create new instance of Manager.
type Opts struct {
	// IncomingEvents is a channel to receive events from.
	IncomingEvents <-chan domain.Event

	// TelegramAPI is a client to interact with telegram API.
	TelegramAPI *tgbotapi.BotAPI
}

// NewManager returns new instance of NewManager.
func NewManager(log *zap.Logger, opts *Opts) *Manager {
	return &Manager{
		log:            log,
		incomingEvents: opts.IncomingEvents,
		eventHandlers:  make(map[int64]chan<- domain.Event),
		telegramAPI:    opts.TelegramAPI,
	}
}

// Run method starts the main goroutine of Manager.
// The call is blocking.
func (mgr *Manager) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-mgr.incomingEvents:
			if !ok {
				return
			}

			// TODO: support concurrent events handling

			switch e := event.(type) {
			case *domain.StartEvent:
				// Check if client already exists
				if _, ok := mgr.eventHandlers[e.ChatID]; ok {
					continue
				}

				// Create and run events handler for new client
				handlerCh := make(chan domain.Event, 1)
				evHandler := handler.NewEventsHandler(mgr.log, &handler.Opts{
					ChatID:         e.ChatID,
					IncomingEvents: handlerCh,
					TelegramAPI:    mgr.telegramAPI,
				})
				go evHandler.Run(ctx)

				// Update mapping and send the event to the newly created handler
				mgr.eventHandlers[e.ChatID] = handlerCh
				handlerCh <- e
			case *domain.LoginEvent:
				// TODO: check if the client is already logged in
				handlerCh, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				handlerCh <- event
			case *domain.ReplyEvent:
				handlerCh, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				handlerCh <- event
			}
		}
	}
}
