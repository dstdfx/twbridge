package manager //nolint

import (
	"context"
	"sync"
	"testing"

	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/handler/mocks"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestManager(t *testing.T) {
	testChatID := int64(123)
	testUserName := "test-user"

	t.Run("handle start event, already has events handler", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleStartEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.StartEvent{
			ChatID:   testChatID,
			FromUser: testUserName,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleStartEvent", mock.Anything)
	})

	t.Run("handle login event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("IsLoggedIn").Return(false)
		eventsHandlerMock.On("HandleLoginEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.LoginEvent{
			ChatID:   testChatID,
			FromUser: testUserName,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "IsLoggedIn")
		eventsHandlerMock.AssertCalled(t, "HandleLoginEvent", mock.Anything)
	})

	t.Run("handle repeated login event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("IsLoggedIn").Return(true)
		eventsHandlerMock.On("HandleRepeatedLoginEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.LoginEvent{
			ChatID:   testChatID,
			FromUser: testUserName,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "IsLoggedIn")
		eventsHandlerMock.AssertCalled(t, "HandleRepeatedLoginEvent", mock.Anything)
	})

	t.Run("handle logout event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleLogoutEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.LogoutEvent{
			ChatID:   testChatID,
			FromUser: testUserName,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleLogoutEvent", mock.Anything)
	})

	t.Run("handle logout event, not logged in", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleLogoutEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.LogoutEvent{
			ChatID:   testChatID,
			FromUser: testUserName,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleLogoutEvent", mock.Anything)
	})

	t.Run("handle help event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleHelpEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.HelpEvent{
			ChatID:   testChatID,
			FromUser: testUserName,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleHelpEvent", mock.Anything)
	})

	t.Run("handle reply event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleReplyEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.ReplyEvent{
			ChatID:    testChatID,
			FromUser:  testUserName,
			Reply:     "test-reply",
			RemoteJid: "test-remote-jid",
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleReplyEvent", mock.Anything)
	})

	t.Run("handle text message event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleTextMessageEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.TextMessageEvent{
			ChatID:             testChatID,
			WhatsappRemoteJid:  "test-remote-jid",
			WhatsappSenderName: "test-sender-name",
			Text:               "test-text",
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleTextMessageEvent", mock.Anything)
	})

	t.Run("handle disconnect event", func(t *testing.T) {
		incomingEventsCh := make(chan domain.Event)
		testMgr := NewManager(zap.NewNop(), &Opts{
			IncomingEvents: incomingEventsCh,
		})

		eventsHandlerMock := &mocks.EventsHandler{}
		eventsHandlerMock.On("HandleDisconnectEvent", mock.Anything).Return(nil)

		// Add test events handler
		testMgr.eventHandlers[testChatID] = eventsHandlerMock

		// Run clients manager in a separate goroutine
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMgr.Run(ctx)
		}()

		// Send login event
		incomingEventsCh <- &domain.DisconnectEvent{
			ChatID: testChatID,
		}

		// Stop clients manager
		cancel()
		wg.Wait()

		eventsHandlerMock.AssertCalled(t, "HandleDisconnectEvent", mock.Anything)
	})
}
