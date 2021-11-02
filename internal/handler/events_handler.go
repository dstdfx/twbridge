package handler

import (
	"bytes"
	"context"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
	whatsappevents "github.com/dstdfx/twbridge/internal/whatsapp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

type EventsHandler struct {
	log          *zap.Logger
	chatID       int64
	eventsCh     chan domain.Event
	telegramAPI  *tgbotapi.BotAPI
	whatsappConn *whatsapp.Conn
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
				eh.handleLoginEvent(e)
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

func (eh *EventsHandler) handleLoginEvent(event *domain.LoginEvent) {
	eh.log.Debug("handle whatsapp login",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	wac, err := whatsapp.NewConnWithOptions(&whatsapp.Options{
		Timeout:         20 * time.Second,
	})
	if err != nil {
		eh.log.Error("failed to establish new whatsapp connection", zap.Error(err))

		return
	}

	eh.whatsappConn = wac

	waHandler := whatsappevents.NewEventsProvider(eh.log, &whatsappevents.Opts{
		OutgoingEvents: eh.eventsCh,
		WhatsappConn: eh.whatsappConn,
	})
	wac.AddHandler(waHandler)
	wac.SetClientVersion(2, 2134, 10)

	qr := make(chan string)
	go func() {
		qrCode, err := qrcode.New(<-qr, qrcode.Low)
		if err != nil {
			eh.log.Error("failed to receive a QR code", zap.Error(err))

			return
		}

		rawCode, err := qrCode.PNG(256)
		if err != nil {
			eh.log.Error("failed to parse QR code", zap.Error(err))

			return
		}

		qrCodeReader := tgbotapi.FileReader{
			Name: "QrCode",
			Reader: bytes.NewReader(rawCode),
			Size: int64(len(rawCode)),
		}

		photo := tgbotapi.NewPhotoUpload(event.ChatID, qrCodeReader)
		if _, err := eh.telegramAPI.Send(photo); err != nil {
			eh.log.Error("failed to send new whatsapp connection", zap.Error(err))

			return
		}


		eh.log.Debug("QR code has been sent")
	}()

	// TODO: save and restore sessions

	session, err := wac.Login(qr)
	if err != nil {
		eh.log.Error("failed to login to whatsapp", zap.Error(err))

		return
	}

	eh.log.Debug("login successful", zap.String("client_id", session.ClientId))

	// TODO: notify via telegram
}
