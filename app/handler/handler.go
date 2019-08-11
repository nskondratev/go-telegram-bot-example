package handler

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nskondratev/go-telegram-translator-bot/app/users"
	"github.com/nskondratev/go-telegram-translator-bot/bot"
	"github.com/rs/zerolog"
)

type Handler struct {
	bot bot.Bot
}

func NewHandler(bot bot.Bot) Handler {
	return Handler{bot: bot}
}

func (h Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	user := ctx.Value("user").(users.User)
	logger := zerolog.Ctx(ctx)

	logger.Info().
		Str("userFromCtx", fmt.Sprintf("%#v", user)).
		Msg("log user in handler")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyToMessageID = update.Message.MessageID

	if _, err := h.bot.Send(msg); err != nil {
		logger.Error().
			Err(err).
			Msg("error while trying to send reply message")
	}
}
