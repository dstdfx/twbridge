package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/dstdfx/twbridge/internal/log"
	"github.com/dstdfx/twbridge/internal/manager"
	"github.com/dstdfx/twbridge/internal/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

// Variables that are injected in build time.
var (
	buildGitCommit string
	buildGitTag    string
	buildDate      string
	buildCompiler  = runtime.Version()
)

const telegramAPITokenEnv = "TELEGRAM_API_TOKEN"

func Start() {
	logger, err := log.NewLogger(zap.DebugLevel, zap.String("service", "twproxy"))
	if err != nil {
		panic(err)
	}

	apiToken, ok := os.LookupEnv(telegramAPITokenEnv)
	if !ok {
		logger.Panic(fmt.Sprintf("%s is required", telegramAPITokenEnv))
	}

	// Create telegram bot instance
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		logger.Panic("failed to run telegram bot", zap.Error(err))
	}

	// Handle interrupt signals
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create telegram events provider instance
	eventsProvider := telegram.NewEventsProvider(logger, &telegram.Opts{
		TelegramAPI: bot,
	})

	// Create clients manager instance
	clientManager := manager.NewManager(logger, &manager.Opts{
		IncomingEvents: eventsProvider.EventsStream(),
		TelegramAPI:    bot,
	})

	go clientManager.Run(rootCtx)

	go func() {
		if err := eventsProvider.Run(rootCtx); err != nil {
			logger.Panic("failed to run telegram events provider", zap.Error(err))
		}
	}()

	select {
	case <-rootCtx.Done():
		stop()
	}
}
