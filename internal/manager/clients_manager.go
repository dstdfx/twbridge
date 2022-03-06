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
	incomingEvents chan domain.Event
	telegramAPI    *tgbotapi.BotAPI
	eventHandlers  map[int64]domain.EventsHandler
}

// Opts represents options to create new instance of Manager.
type Opts struct {
	// IncomingEvents is a channel to receive events from.
	IncomingEvents chan domain.Event

	// TelegramAPI is a client to interact with telegram API.
	TelegramAPI *tgbotapi.BotAPI
}

// NewManager returns new instance of NewManager.
func NewManager(log *zap.Logger, opts *Opts) *Manager {
	return &Manager{
		log:            log,
		incomingEvents: opts.IncomingEvents,
		eventHandlers:  make(map[int64]domain.EventsHandler),
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
				var eventsHandler domain.EventsHandler

				// Check if the client already has an events handler
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					// Create events handler for new client and handle event
					eventsHandler = handler.NewEventsHandler(mgr.log, &handler.Opts{
						ChatID:                 e.ChatID,
						WhatsappProviderEvents: mgr.incomingEvents,
						TelegramAPI:            mgr.telegramAPI,
					})

					// Add it to the mapping
					mgr.eventHandlers[e.ChatID] = eventsHandler
				}

				// Handle start event
				if err := eventsHandler.HandleStartEvent(e); err != nil {
					mgr.log.Error("failed to handle start event", zap.Error(err))
				}
			case *domain.LoginEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if eventsHandler.IsLoggedIn() {
					if err := eventsHandler.HandleRepeatedLoginEvent(e); err != nil {
						mgr.log.Error("failed to handle repeated login event", zap.Error(err))
					}

					continue
				}

				if err := eventsHandler.HandleLoginEvent(e); err != nil {
					mgr.log.Error("failed to handle login event", zap.Error(err))
				}
			case *domain.LogoutEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if err := eventsHandler.HandleLogoutEvent(e); err != nil {
					mgr.log.Error("failed to handle logout event", zap.Error(err))
				}
			case *domain.HelpEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if err := eventsHandler.HandleHelpEvent(e); err != nil {
					mgr.log.Error("failed to handle help event", zap.Error(err))
				}
			case *domain.ReplyEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if err := eventsHandler.HandleReplyEvent(e); err != nil {
					mgr.log.Error("failed to handle reply event", zap.Error(err))
				}
			case *domain.TextMessageEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if err := eventsHandler.HandleTextMessageEvent(e); err != nil {
					mgr.log.Error("failed to handle text message event", zap.Error(err))
				}
			case *domain.ImageMessageEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if err := eventsHandler.HandleImageMessageEvent(e); err != nil {
					mgr.log.Error("failed to handle text message event", zap.Error(err))
				}
			case *domain.DisconnectEvent:
				eventsHandler, ok := mgr.eventHandlers[e.ChatID]
				if !ok {
					mgr.log.Error("failed to find events handler for the chat_id",
						zap.Int64("chat_id", e.ChatID))

					continue
				}

				if err := eventsHandler.HandleDisconnectEvent(e); err != nil {
					mgr.log.Error("failed to handle disconnect event", zap.Error(err))
				}
			}
		}
	}
}
