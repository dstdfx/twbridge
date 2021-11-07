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

const (
	telegramAPITokenEnv = "TELEGRAM_API_TOKEN"

	defaultTelegramReceiveTimeout = 60
)

func Start() {
	logger, err := log.NewLogger(zap.DebugLevel, zap.String("service", "twbridge"))
	if err != nil {
		panic(err)
	}

	apiToken, ok := os.LookupEnv(telegramAPITokenEnv)
	if !ok {
		logger.Panic(fmt.Sprintf("%s is required", telegramAPITokenEnv))
	}

	logger.Info("twbridge is running...",
		zap.String("build_commit", buildGitCommit),
		zap.String("build_tag", buildGitTag),
		zap.String("build_date", buildDate),
		zap.String("go_version", buildCompiler))

	// Create telegram bot instance
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		logger.Panic("failed to run telegram bot", zap.Error(err))
	}

	// Handle interrupt signals
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// TODO: use webhook for receiving tg updates

	// Create telegram config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = defaultTelegramReceiveTimeout

	// Create telegram updates channel
	tgUpdatesCh, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Panic("failed to get updates chan", zap.Error(err))
	}
	defer bot.StopReceivingUpdates()

	// Create telegram events provider instance
	eventsProvider := telegram.NewEventsProvider(logger, &telegram.Opts{
		TelegramUpdates: tgUpdatesCh,
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
