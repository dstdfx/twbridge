package handler_test

import (
	"testing"

	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/handler"
	"go.uber.org/zap"
)

func TestEventsHandler(t *testing.T) {

	testChatID := int64(123)
	eventsHandler:= handler.NewEventsHandler(zap.NewNop(), &handler.Opts{
		ChatID:                 testChatID,
	})


	t.Run("handle start event", func(t *testing.T) {

	})

	t.Run("handle login event", func(t *testing.T) {

	})

	t.Run("handle reply event", func(t *testing.T) {

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

		//Mock reply event
		replyEvent := domain.ReplyEvent{
			ChatID: testChatID,
			FromUser: "User1",
			Reply: "User2",
			RemoteJid: "123",
		}

		eventsHandler.HandleReplyEvent(e); err != nil {

	})

	t.Run("handle text message event", func(t *testing.T) {

	})
}
