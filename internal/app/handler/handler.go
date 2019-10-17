package handler

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"

	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
)

type Handler struct {
	bot bot.Bot
}

func NewHandler(bot bot.Bot) Handler {
	return Handler{bot: bot}
}

func (h Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	user := users.Ctx(ctx)
	logger := zerolog.Ctx(ctx)

	logger.Info().
		Str("userFromCtx", fmt.Sprintf("%#v", user)).
		Msg("log user in handler")

	if update.Message != nil && update.Message.Voice != nil {
		v := update.Message.Voice
		logger.Info().
			Int("duration", v.Duration).
			Str("file_id", v.FileID).
			Str("mimetype", v.MimeType).
			Int("duration", v.Duration).
			Msg("Incoming voice message")
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyToMessageID = update.Message.MessageID

	if _, err := h.bot.Send(msg); err != nil {
		logger.Error().
			Err(err).
			Msg("error while trying to send reply message")
	}
}
