package command

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"
	"golang.org/x/text/language"

	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
)

type Handler struct {
	bot             *bot.Bot
	us              users.Store
	newUsersMetric  *metrics.Counter
	commandsLatency *metrics.Latency
}

func New(bot *bot.Bot, us users.Store) *Handler {
	return &Handler{
		bot: bot,
		us:  us,
		newUsersMetric: metrics.NewCounter(
			"bot_command_handler_new_users_count",
			"Count of users, that execute start command",
			[]string{"lang"},
		),
		commandsLatency: metrics.NewLatency(
			"bot_command_handler_latency",
			"Latency for handling time by command and status",
			nil,
			[]string{"command"},
		),
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
	cmd := update.Message.Command()
	switch cmd {
	case "start":
		h.onStart(ctx, update)
	case "help":
		h.onHelp(ctx, update)
	case "lang":
		h.onLang(ctx, update)
	case "toggle":
		h.onToggle(ctx, update)
	default:
		h.onUnknown(ctx, update)
	}
}

func (h *Handler) onStart(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	l := h.commandsLatency.NewAction("start")
	user := users.Ctx(ctx)
	m := h.newUsersMetric.NewAction(user.Lang)
	m.Inc(metrics.StatusOk)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! You can record voice messages and get voice translation for them")
	if _, err := h.bot.Send(msg); err != nil {
		log.Error().
			Err(err).
			Msg("error while trying to send reply message")
		l.Observe(metrics.StatusErr)
		return
	}
	l.Observe(metrics.StatusOk)
}

func (h *Handler) onHelp(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	l := h.commandsLatency.NewAction("help")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Help command is received")
	if _, err := h.bot.Send(msg); err != nil {
		log.Error().
			Err(err).
			Msg("error while trying to send reply message")
		l.Observe(metrics.StatusErr)
		return
	}
	l.Observe(metrics.StatusOk)
}

func (h *Handler) onToggle(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	l := h.commandsLatency.NewAction("toggle")
	user := users.Ctx(ctx)
	if user != nil {
		if err := h.us.UpdateTranslationLangs(ctx, user.TelegramUserID, user.TargetLang, user.SourceLang); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to update translation languages")
			l.Observe(metrics.StatusErr)
			return
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Languages are changed.")
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
			l.Observe(metrics.StatusErr)
			return
		}
		l.Observe(metrics.StatusOk)
	}
}

func (h *Handler) onLang(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	l := h.commandsLatency.NewAction("lang")
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
			l.Observe(metrics.StatusErr)
			return
		}
		l.Observe(metrics.StatusOk)
		return
	}
	if _, err := language.Parse(args[0]); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You should provide correct language, for example: \"en\"")
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
			l.Observe(metrics.StatusErr)
			return
		}
		l.Observe(metrics.StatusOk)
		return
	}
	if _, err := language.Parse(args[1]); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You should provide correct language, for example: \"en\"")
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
			l.Observe(metrics.StatusErr)
			return
		}
		l.Observe(metrics.StatusOk)
		return
	}
	user := users.Ctx(ctx)
	if user != nil {
		if err := h.us.UpdateTranslationLangs(ctx, user.TelegramUserID, args[0], args[1]); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to update translation languages")
			l.Observe(metrics.StatusErr)
			return
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Languages are changed.")
		if _, err := h.bot.Send(msg); err != nil {
			log.Error().
				Err(err).
				Msg("error while trying to send reply message")
			l.Observe(metrics.StatusErr)
			return
		}
	}
	l.Observe(metrics.StatusOk)
}

func (h *Handler) onUnknown(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	l := h.commandsLatency.NewAction(update.Message.Command())
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command.")
	if _, err := h.bot.Send(msg); err != nil {
		log.Error().
			Err(err).
			Msg("error while trying to send reply message")
		l.Observe(metrics.StatusErr)
		return
	}
	l.Observe(metrics.StatusOk)
}
