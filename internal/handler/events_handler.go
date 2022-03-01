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

const disconnectMsg = `The session is invalidated due to internal error, please repeat login process again.`

const helpMsg = `
Supported commands:
/start - prints starting message
/login - establishes a session with WhatsApp via QR-code challenge
/logout - invalidates your current session with WhatsApp
/help - prints this message
`

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

	if err := eh.notifyTelegram(startMsg); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
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
			eh.log.Error("failed to receive QR-code", zap.Error(err))

			return
		}

		rawCode, err := qrCode.PNG(defaultQRCodePNGSize)
		if err != nil {
			eh.log.Error("failed to parse QR-code", zap.Error(err))

			return
		}

		qrCodeReader := tgbotapi.FileReader{
			Name:   "QrCode",
			Reader: bytes.NewReader(rawCode),
			Size:   int64(len(rawCode)),
		}

		photo := tgbotapi.NewPhotoUpload(eh.chatID, qrCodeReader)
		if _, err := eh.telegramAPI.Send(photo); err != nil {
			eh.log.Error("failed to send QR-code", zap.Error(err))

			return
		}

		eh.log.Debug("QR-code has been sent")
	}()

	// TODO: save and restore sessions

	session, err := wac.Login(qr)
	if err != nil {
		err := eh.notifyTelegram("QR-code scanning timed out, let's try again, type /login")
		if err != nil {
			return fmt.Errorf("failed to notify telegram: %w", err)
		}

		return fmt.Errorf("failed to login to whatsapp: %w", err)
	}

	eh.mu.Lock()
	eh.isWhatsAppLoggedIn = true
	eh.mu.Unlock()

	eh.log.Debug("login successful", zap.String("client_id", session.ClientId))

	if err := eh.notifyTelegram("Successfully logged in"); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
	}

	return nil
}

// HandleLogoutEvent method handles repeated logout event.
func (eh *EventsHandler) HandleLogoutEvent(event *domain.LogoutEvent) error {
	eh.log.Debug("handle whatsapp logout",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	// Check if the client is already logged out
	if !eh.IsLoggedIn() {
		if err := eh.notifyTelegram("Already logged out"); err != nil {
			return fmt.Errorf("failed to notify telegram: %w", err)
		}

		return nil
	}

	if err := eh.whatsappClient.Logout(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	eh.mu.Lock()
	eh.isWhatsAppLoggedIn = false
	eh.mu.Unlock()

	if err := eh.notifyTelegram("Successfully logged out"); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
	}

	return nil
}

// HandleRepeatedLoginEvent method handles repeated login event.
func (eh *EventsHandler) HandleRepeatedLoginEvent(event *domain.LoginEvent) error {
	eh.log.Debug("handle repeated login event",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	if err := eh.notifyTelegram("Already logged in"); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
	}

	return nil
}

// HandleHelpEvent method handles help event.
func (eh *EventsHandler) HandleHelpEvent(event *domain.HelpEvent) error {
	eh.log.Debug("handle help event",
		zap.String("username", event.FromUser),
		zap.Int64("chat_id", event.ChatID))

	if err := eh.notifyTelegram(helpMsg); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
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

	if err := eh.notifyTelegram(textMessageTemplate); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
	}

	return nil
}

// HandleImageMessageEvent method handles image message event.
func (eh *EventsHandler) HandleImageMessageEvent(event *domain.ImageMessageEvent) error {

	eh.log.Debug("handle image message event",
		zap.String("remote_jid", event.WhatsappRemoteJid))

	imageReader := tgbotapi.FileReader{
		Name:   "ImageMessage",
		Reader: bytes.NewReader(event.ImageBytes),
		Size:   int64(len(event.ImageBytes)),
	}

	photo := tgbotapi.NewPhotoUpload(eh.chatID, imageReader)
	if _, err := eh.telegramAPI.Send(photo); err != nil {
		eh.log.Error("failed to send QR-code", zap.Error(err))

		return fmt.Errorf("failed to send image message chat_id=%d: %w",
			event.ChatID,
			err)
	}

	eh.log.Debug("Image message from whatsapp has been forwarded")

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

// HandleDisconnectEvent method handles disconnect event.
func (eh *EventsHandler) HandleDisconnectEvent(event *domain.DisconnectEvent) error {
	eh.log.Debug("handle disconnect event",
		zap.Int64("chat_id", event.ChatID))

	// Attempt to logout the client and stop whatsapp handler
	if err := eh.whatsappClient.Logout(); err != nil {
		eh.log.Error("failed to logout disconnected client", zap.Error(err))
	}

	eh.mu.Lock()
	eh.isWhatsAppLoggedIn = false
	eh.mu.Unlock()

	if err := eh.notifyTelegram(disconnectMsg); err != nil {
		return fmt.Errorf("failed to notify telegram: %w", err)
	}

	return nil
}

func (eh *EventsHandler) notifyTelegram(msg string) error {
	if _, err := eh.telegramAPI.Send(tgbotapi.NewMessage(eh.chatID, msg)); err != nil {
		return fmt.Errorf("failed to send message to telegram: %w", err)
	}

	return nil
}
