package handler

import (
	"context"

	"github.com/dstdfx/twbridge/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

type EventsHandler struct {
	log          *zap.Logger
	chatID       int64
	eventsCh     chan domain.Event
	telegramAPI  *tgbotapi.BotAPI
}

type Opts struct {
	ChatID int64
	IncomingEvents chan domain.Event
	TelegramAPI    *tgbotapi.BotAPI
}

func NewEventsHandler(log *zap.Logger, opts *Opts) *EventsHandler {
	return &EventsHandler{
		log:         log,
		chatID: opts.ChatID,
		eventsCh:    opts.IncomingEvents,
		telegramAPI: opts.TelegramAPI,
	}
}

func (eh *EventsHandler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eh.eventsCh:
			if !ok {
				return
			}

			switch e := event.(type) {
			case *domain.StartEvent:
				eh.handleStartEvent(e)
			case *domain.LoginEvent:
				// TODO:
			case *domain.TextMessageEvent:
				// TODO:
			}
		}
	}
}

func (eh *EventsHandler) handleStartEvent(event *domain.StartEvent) {
	eh.log.Debug("handle start",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	msg := tgbotapi.NewMessage(event.ChatID, `
		Hello, this is telegram<->whatsapp bridge that allows you to get your whatsapp messages here.
		To start, you need to scan a QR code that will appear here with whatsapp application on your phone.
		Type /login to begin.
	`)

	if _, err := eh.telegramAPI.Send(msg); err != nil {
		eh.log.Error("failed to send start message to telegram", zap.Error(err))
	}
}
