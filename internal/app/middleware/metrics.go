package middleware

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
)

func PromUpdate(next bot.Handler) bot.Handler {
	m := metrics.NewLatency(
		"bot_update_handle_time_nanoseconds",
		"Latency for handling update time",
		nil,
		[]string{"update_type"},
	)

	return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
		a := m.NewAction(getUpdateType(update))
		next.Handle(ctx, update)
		a.Observe(metrics.StatusOk)
	})
}

func getUpdateType(update tgbotapi.Update) string {
	switch {
	case update.Message != nil && update.Message.IsCommand():
		return "command"
	case update.Message != nil && update.Message.Voice != nil:
		return "voice"
	// TODO: Add more update types
	default:
		return "unknown"
	}
}
