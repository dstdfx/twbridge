package domain

type EventProviderType string

const (
	TelegramEventProvider EventProviderType = "telegram"
	WhatsappEventProvider EventProviderType = "whatsapp"
)

type EventProvider interface {
	EventsStream() <-chan Event
	Type() EventProviderType
}

type EventsHandler interface {
}

type EventType string

const (
	StartEventType EventType = "start"
	LoginEventType EventType = "login"
	TextMessageEventType EventType = "text_message" // whatsapp only
	ReplyEventType EventType = "reply" // telegram only
)

type Event interface {
	Type() EventType
}

type StartEvent struct {
	ChatID int64
	FromUser string
}

func (se *StartEvent) Type() EventType {
	return StartEventType
}

type LoginEvent struct {
	ChatID int64
	FromUser string
}

func (le *LoginEvent) Type() EventType {
	return LoginEventType
}

type TextMessageEvent struct {
	WhatsappSenderJid string
	WhatsappSenderName string
	Text string
}

func (te *TextMessageEvent) Type() EventType {
	return TextMessageEventType
}

type ReplyEvent struct {
	ChatID int64
	FromUser string
	Reply string
	// TODO: add some identifier to know whom to reply
}

func (re *ReplyEvent) Type() EventType {
	return ReplyEventType
}
