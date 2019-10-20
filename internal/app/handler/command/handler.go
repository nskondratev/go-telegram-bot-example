package command

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
	"github.com/rs/zerolog"
	"golang.org/x/text/language"
	"strings"
)

type Handler struct {
	bot *bot.Bot
	us  users.Store
}

func New(bot *bot.Bot, us users.Store) *Handler {
	return &Handler{
		bot: bot,
		us:  us,
	}
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
	case "lang":
		args := strings.Split(update.Message.CommandArguments(), " ")
		log.Info().
			Strs("args", args).
			Msg("lang command arguments")
		if len(args) != 2 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You should provide two languages, for example: \"en ru\"")
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		}
		if _, err := language.Parse(args[0]); err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You should provide correct language, for example: \"en\"")
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		}
		if _, err := language.Parse(args[1]); err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You should provide correct language, for example: \"en\"")
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		}
		user := users.Ctx(ctx)
		if user != nil {
			if err := h.us.UpdateTranslationLangs(ctx, user.TelegramUserID, args[0], args[1]); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to update translation languages")
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Languages are changed.")
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		}
	case "toggle":
		user := users.Ctx(ctx)
		if user != nil {
			if err := h.us.UpdateTranslationLangs(ctx, user.TelegramUserID, user.TargetLang, user.SourceLang); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to update translation languages")
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Languages are changed.")
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		}
	}
}
