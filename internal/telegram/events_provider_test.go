package telegram_test

import (
	"context"
	"sync"
	"testing"

	"go.uber.org/zap"

	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

var _ domain.EventProvider = new(telegram.EventsProvider)

func TestEventsProvider(t *testing.T) {
	tgUpdatesCh := make(chan tgbotapi.Update)
	eventsProvider := telegram.NewEventsProvider(zap.NewNop(), &telegram.Opts{TelegramUpdates: tgUpdatesCh})

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := eventsProvider.Run(rootCtx); err != nil {
			t.Errorf("failed to run events provider: %s", err)
		}
	}()

	t.Run("start event", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			defer wg.Done()
			gotEvent = <-eventsProvider.EventsStream()
		}()

		// Emulate telegram update message
		testUpdate := tgbotapi.Update{
			UpdateID: 1,
			Message: &tgbotapi.Message{
				MessageID: 1,
				From: &tgbotapi.User{
					FirstName: "test name",
					LastName:  "test surname",
					UserName:  "testuser",
				},
				Chat: &tgbotapi.Chat{
					ID: 42,
				},
				Text: "/start",
			},
		}
		tgUpdatesCh <- testUpdate

		// Wait for the event to be processed
		wg.Wait()

		assert.Equal(t, domain.StartEventType, gotEvent.Type())
		gotStartEvent := gotEvent.(*domain.StartEvent)

		assert.Equal(t, testUpdate.Message.Chat.ID, gotStartEvent.ChatID)
		assert.Equal(t, testUpdate.Message.From.UserName, gotStartEvent.FromUser)
	})

	t.Run("login event", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		var gotEvent domain.Event
		go func() {
			defer wg.Done()
			gotEvent = <-eventsProvider.EventsStream()
		}()

		// Emulate telegram update message
		testUpdate := tgbotapi.Update{
			UpdateID: 1,
			Message: &tgbotapi.Message{
				MessageID: 1,
				From: &tgbotapi.User{
					FirstName: "test name",
					LastName:  "test surname",
					UserName:  "testuser",
				},
				Chat: &tgbotapi.Chat{
					ID: 42,
				},
				Text: "/login",
			},
		}
		tgUpdatesCh <- testUpdate

		// Wait for the event to be processed
		wg.Wait()

		assert.Equal(t, domain.LoginEventType, gotEvent.Type())
		gotStartEvent := gotEvent.(*domain.LoginEvent)

		assert.Equal(t, testUpdate.Message.Chat.ID, gotStartEvent.ChatID)
		assert.Equal(t, testUpdate.Message.From.UserName, gotStartEvent.FromUser)
	})

	t.Run("ignored message", func(t *testing.T) {
		// Emulate telegram update message that will be ignored
		testUpdate := tgbotapi.Update{
			UpdateID: 1,
			Message:  nil,
		}
		tgUpdatesCh <- testUpdate

		eventsCh := eventsProvider.EventsStream()
		assert.Len(t, eventsCh, 0)
	})
}
