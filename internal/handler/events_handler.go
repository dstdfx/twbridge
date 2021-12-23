package handler

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/dstdfx/twbridge/internal/domain"
	whatsappevents "github.com/dstdfx/twbridge/internal/whatsapp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

const (
	defaultWhatsappClientMajorVersion = 2
	defaultWhatsappClientMinorVersion = 2134
	defaultWhatsappClientPatchVersion = 10
	defaultWhatsappConnTimeout        = 20 * time.Second

	defaultQRCodePNGSize = 256
)

const startMsg = `
Hi, this is Telegram<->WhatsApp bridge that allows you to receive your WhatsApp messages here and also reply to them.

All you need is to scan a QR-code that will appear here after you click on /login.
Use WhatsApp application on your phone to scan it (Settings -> Linked Devices -> Link a Device).
This step is needed in order to authenticate you in WhatsApp.

So let's get it started.`

// EventsHandler represents entity that handles events from telegram and whatsapp
// event providers.
type EventsHandler struct {
	log                *zap.Logger
	chatID             int64
	eventsCh           chan domain.Event
	telegramAPI        *tgbotapi.BotAPI
	whatsappClient     domain.WhatsappClient
	mu                 sync.RWMutex
	isWhatsAppLoggedIn bool
}

// Opts represents options to create new instance of EventsHandler.
type Opts struct {
	// ChatID is telegram bot chat identifier.
	ChatID int64

	// WhatsappProviderEvents is a channel to send events from whatsapp provider.
	WhatsappProviderEvents chan domain.Event

	// TelegramAPI is a client to interact with telegram API.
	TelegramAPI *tgbotapi.BotAPI
}

// NewEventsHandler creates new instance of EventsHandler.
func NewEventsHandler(log *zap.Logger, opts *Opts) *EventsHandler {
	return &EventsHandler{
		log:         log,
		chatID:      opts.ChatID,
		eventsCh:    opts.WhatsappProviderEvents,
		telegramAPI: opts.TelegramAPI,
	}
}

// IsLoggedIn method returns `true` if client is authenticated in WhatsApp, otherwise
// returns `false`.
func (eh *EventsHandler) IsLoggedIn() bool {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	return eh.isWhatsAppLoggedIn
}

// HandleStartEvent method handles start event.
func (eh *EventsHandler) HandleStartEvent(event *domain.StartEvent) error {
	eh.log.Debug("handle start",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	msg := tgbotapi.NewMessage(event.ChatID, startMsg)

	if _, err := eh.telegramAPI.Send(msg); err != nil {
		return fmt.Errorf("failed to send start message to telegram: %w", err)
	}

	return nil
}

// HandleLoginEvent method handles login event.
func (eh *EventsHandler) HandleLoginEvent(event *domain.LoginEvent) error {
	eh.log.Debug("handle whatsapp login",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	wac, err := whatsapp.NewConnWithOptions(&whatsapp.Options{
		Timeout: defaultWhatsappConnTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to establish new whatsapp connection: %w", err)
	}

	// Initialize new whatsapp client
	eh.whatsappClient = whatsappevents.NewClient(wac)

	// Initialize whatsapp events provider
	waHandler := whatsappevents.NewEventsProvider(eh.log, &whatsappevents.Opts{
		ChatID:         eh.chatID,
		OutgoingEvents: eh.eventsCh,
		WhatsappClient: eh.whatsappClient,
	})
	wac.AddHandler(waHandler)
	wac.SetClientVersion(
		defaultWhatsappClientMajorVersion,
		defaultWhatsappClientMinorVersion,
		defaultWhatsappClientPatchVersion)

	qr := make(chan string)
	go func() {
		qrCode, err := qrcode.New(<-qr, qrcode.Low)
		if err != nil {
			eh.log.Error("failed to receive a QR code", zap.Error(err))

			return
		}

		rawCode, err := qrCode.PNG(defaultQRCodePNGSize)
		if err != nil {
			eh.log.Error("failed to parse QR code", zap.Error(err))

			return
		}

		qrCodeReader := tgbotapi.FileReader{
			Name:   "QrCode",
			Reader: bytes.NewReader(rawCode),
			Size:   int64(len(rawCode)),
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
		// TODO: move to a separate error handling fn
		msg := tgbotapi.NewMessage(eh.chatID, "QR-code scanning timed out, let's try again, type /login")
		if _, err := eh.telegramAPI.Send(msg); err != nil {
			eh.log.Error("failed to send message to telegram", zap.Error(err))
		}

		return fmt.Errorf("failed to login to whatsapp: %w", err)
	}

	eh.mu.Lock()
	eh.isWhatsAppLoggedIn = true
	eh.mu.Unlock()

	eh.log.Debug("login successful", zap.String("client_id", session.ClientId))

	msg := tgbotapi.NewMessage(eh.chatID, "Successfully logged in")
	if _, err := eh.telegramAPI.Send(msg); err != nil {
		return fmt.Errorf("failed to send message to telegram: %w", err)
	}

	return nil
}

// HandleLogoutEvent method handles repeated logout event.
func (eh *EventsHandler) HandleLogoutEvent(event *domain.LogoutEvent) error {
	eh.log.Debug("handle whatsapp logout",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	if err := eh.whatsappClient.Logout(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	eh.mu.Lock()
	eh.isWhatsAppLoggedIn = false
	eh.mu.Unlock()

	msg := tgbotapi.NewMessage(eh.chatID, "Successfully logged out")
	if _, err := eh.telegramAPI.Send(msg); err != nil {
		return fmt.Errorf("failed to send message to telegram: %w", err)
	}

	return nil
}

// HandleRepeatedLoginEvent method handles repeated login event.
func (eh *EventsHandler) HandleRepeatedLoginEvent(event *domain.LoginEvent) error {
	eh.log.Debug("handle repeated login event",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	msg := tgbotapi.NewMessage(eh.chatID, "Already logged in")
	if _, err := eh.telegramAPI.Send(msg); err != nil {
		return fmt.Errorf("failed to send message to telegram: %w", err)
	}

	return nil
}

// HandleTextMessageEvent method handles text message event.
func (eh *EventsHandler) HandleTextMessageEvent(event *domain.TextMessageEvent) error {
	eh.log.Debug("handle text message event",
		zap.String("remote_jid", event.WhatsappRemoteJid))

	textMessageTemplate := fmt.Sprintf(domain.TextMessageFmt,
		event.WhatsappSenderName,
		event.WhatsappRemoteJid,
		event.Text)

	msg := tgbotapi.NewMessage(eh.chatID, textMessageTemplate)
	if _, err := eh.telegramAPI.Send(msg); err != nil {
		return fmt.Errorf("failed to send message to telegram: %w", err)
	}

	return nil
}

// HandleReplyEvent method handles reply event.
func (eh *EventsHandler) HandleReplyEvent(event *domain.ReplyEvent) error {
	eh.log.Debug("reply to a message",
		zap.Int64("chat_id", event.ChatID),
		zap.String("remote_jid", event.RemoteJid))

	msg := &domain.WhatsappTextMessage{
		RemoteJid: event.RemoteJid,
		Text:      event.Reply,
	}

	if err := eh.whatsappClient.Send(msg); err != nil {
		return fmt.Errorf("failed to send message chat_id=%d remote_jid=%s: %w",
			event.ChatID,
			event.RemoteJid,
			err)
	}

	return nil
}
