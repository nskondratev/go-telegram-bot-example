package callbackquery

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/lang"
	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
	"github.com/rs/zerolog"
	"strings"
)

type Handler struct {
	bot                    *bot.Bot
	us                     users.Store
	warnErrorsMetric       *metrics.Counter
	sendReplyLatencyMetric *metrics.Latency
}

func New(bot *bot.Bot, us users.Store) *Handler {
	return &Handler{
		bot: bot,
		us:  us,
	}
}

func (h *Handler) Middleware(next bot.Handler) bot.Handler {
	return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
		if update.CallbackQuery != nil {
			h.Handle(ctx, update)
			return
		}
		next.Handle(ctx, update)
	})
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	tokens := strings.Split(update.CallbackQuery.Data, ":")
	log.Info().
		Str("data", update.CallbackQuery.Data).
		Strs("tokens", tokens).
		Msg("callback query update handler")
	if len(tokens) > 0 {
		switch tokens[0] {
		case "source_lang":
			h.onSourceLang(ctx, update)
		case "target_lang":
			h.onTargetLang(ctx, update)
		}
	}
}

func (h *Handler) onSourceLang(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	user := users.Ctx(ctx)
	if user != nil {
		sourceLang := getLangValue(update.CallbackQuery.Data)
		log.Info().
			Str("username", user.UserName).
			Str("data", update.CallbackQuery.Data).
			Str("lang", sourceLang).
			Msg("on source lang callback query handler")
		if err := h.us.UpdateSourceLang(ctx, user.TelegramUserID, sourceLang); err != nil {
			log.Error().
				Err(err).
				Str("username", user.UserName).
				Int64("telegram_user_id", user.TelegramUserID).
				Str("data", update.CallbackQuery.Data).
				Str("lang", sourceLang).
				Msg("Failed to update source language for user")
			return
		}
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"Great! Now choose the second language for translation:",
		)
		msg.ReplyMarkup = &lang.TargetLanguagesKeyboard
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
		}
	} else {
		log.Warn().Msg("User is nil :(")
	}
}

func (h *Handler) onTargetLang(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	user := users.Ctx(ctx)
	if user != nil {
		targetLang := getLangValue(update.CallbackQuery.Data)
		log.Info().
			Str("username", user.UserName).
			Str("data", update.CallbackQuery.Data).
			Str("lang", targetLang).
			Msg("on target lang callback query handler")
		if err := h.us.UpdateTargetLang(ctx, user.TelegramUserID, targetLang); err != nil {
			log.Error().
				Err(err).
				Str("username", user.UserName).
				Int64("telegram_user_id", user.TelegramUserID).
				Str("data", update.CallbackQuery.Data).
				Str("lang", targetLang).
				Msg("Failed to update target language for user")
			return
		}
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"Great, you've changed your translation languages.",
		)
		msg.ReplyMarkup = nil
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
		}
	} else {
		log.Warn().Msg("User is nil :(")
	}
}

func getLangValue(data string) string {
	ts := strings.Split(data, ":")
	if len(ts) > 1 {
		return ts[1]
	}
	return ""
}
