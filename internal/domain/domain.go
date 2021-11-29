package domain

// TextMessageFmt represents a message format that will be sent to a user in
// case incoming text messages from whatsapp.
const TextMessageFmt = "From: %s [jid: %s] \n= = = = = = = = = = = =\nMessage: %s"

// EventType represents an event type.
type EventType string

const (
	StartEventType       EventType = "start"
	LoginEventType       EventType = "login"
	TextMessageEventType EventType = "text_message" // whatsapp only
	ReplyEventType       EventType = "reply"        // telegram only
)

// Event represents a generic event API.
type Event interface {
	Type() EventType
}

// StartEvent represents an initial event.
type StartEvent struct {
	// ChatID is telegram bot chat identifier.
	ChatID int64

	// FromUser is a telegram username of the client that interacts with the bot.
	FromUser string
}

func (se *StartEvent) Type() EventType {
	return StartEventType
}

// LoginEvent represents a login event.
type LoginEvent struct {
	// ChatID is telegram bot chat identifier.
	ChatID int64

	// FromUser is a telegram username of the client that interacts with the bot.
	FromUser string
}

func (le *LoginEvent) Type() EventType {
	return LoginEventType
}

// TextMessageEvent represents an incoming text message event.
type TextMessageEvent struct {
	// WhatsappRemoteJid is a whatsapp client identifier that sent the message.
	WhatsappRemoteJid string

	// WhatsappSenderName is a whatsapp client's name that sent the message.
	WhatsappSenderName string

	// Text is a text message body.
	Text string
}

func (te *TextMessageEvent) Type() EventType {
	return TextMessageEventType
}

// ReplyEvent represents a message reply event.
type ReplyEvent struct {
	// ChatID is telegram bot chat identifier.
	ChatID int64

	// FromUser is a telegram username of the client that interacts with the bot.
	FromUser string

	// Reply is a reply text message body.
	Reply string

	// RemoteJid is a whatsapp user identifier.
	RemoteJid string
}

func (re *ReplyEvent) Type() EventType {
	return ReplyEventType
}

/* Whatsapp related domain entities */

// WhatsappContact represents whatsapp contact.
type WhatsappContact struct {
	// Jid is a contact's identifier.
	Jid string

	// Name is a name of the contact.
	Name string
}

// WhatsappMessageType represents whatsapp message type.
type WhatsappMessageType string

const WhatsappTextMessageType = "text_message"

// WhatsappMessage is an interface that represents whatsapp messages in general.
type WhatsappMessage interface {
	Type() WhatsappMessageType
}

// WhatsappTextMessage represents a whatsapp text message.
type WhatsappTextMessage struct {
	// RemoteJid is an identifier of a user the message is sent to.
	RemoteJid string

	// Text is a text of the message.
	Text string
}

// Type method returns type of the message.
func (msg *WhatsappTextMessage) Type() WhatsappMessageType {
	return WhatsappTextMessageType
}

// WhatsappClient represents a common interface that describes whatsapp client behaviour.
type WhatsappClient interface {
	Restore() error
	GetContacts() map[string]WhatsappContact
	Send(msg WhatsappMessage) error
}
