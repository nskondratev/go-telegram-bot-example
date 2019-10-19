package voice

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"

	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
	"github.com/nskondratev/go-telegram-translator-bot/internal/voicetranslate"
)

type Handler struct {
	bot   *bot.Bot
	voice *voicetranslate.Service
}

func New(bot *bot.Bot, voice *voicetranslate.Service) *Handler {
	return &Handler{
		bot:   bot,
		voice: voice,
	}
}

func (h *Handler) Middleware(next bot.Handler) bot.Handler {
	return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
		if update.Message != nil && update.Message.Voice != nil {
			h.Handle(ctx, update)
			return
		}
		next.Handle(ctx, update)
	})
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	log := zerolog.Ctx(ctx)
	if update.Message != nil && update.Message.Voice != nil {
		v := update.Message.Voice
		fileURL, err := h.bot.GetFileDirectURL(v.FileID)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to get file direct URL")
			return
		}
		data, err := getBytesFromURL(fileURL)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to get bytes from file URL")
			return
		}

		user := users.Ctx(ctx)
		genSpeech, err := h.voice.Translate(ctx, data, user.SourceLang, user.TargetLang)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to translate voice message")
			return
		}

		if !genSpeech.UseExisting() {
			msg := tgbotapi.NewVoiceUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Bytes: genSpeech.Voice})
			msg.ReplyToMessageID = update.Message.MessageID
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		}
	}
}

func getBytesFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to make get request: %w", err)
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bs, fmt.Errorf("failed to read response body: %w", err)
	}
	return bs, nil
}
