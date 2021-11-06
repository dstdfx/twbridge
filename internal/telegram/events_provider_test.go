package telegram_test

import (
	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/dstdfx/twbridge/internal/telegram"
)

var _ domain.EventProvider = new(telegram.EventsProvider)

// TODO: to be added
