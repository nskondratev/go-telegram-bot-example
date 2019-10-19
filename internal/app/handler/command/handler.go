package command

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/rs/zerolog"
)

type Handler struct {
	bot *bot.Bot
}

func New(bot *bot.Bot) *Handler {
	return &Handler{bot: bot}
}

func (h *Handler) Middleware(next bot.Handler) bot.Handler {
	return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
		if update.Message != nil && update.Message.IsCommand() {
			h.Handle(ctx, update)
			return
		}
		next.Handle(ctx, update)
	})
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	switch update.Message.Command() {
	case "help":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Help command is received")
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
		}
	}
}
