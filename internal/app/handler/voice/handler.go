package voice

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"

	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
	"github.com/nskondratev/go-telegram-translator-bot/internal/voicetranslate"
)

type usersCostCharger interface {
	ChargeCost(ctx context.Context, tgUserID, cost int64) error
}

type Handler struct {
	bot                        *bot.Bot
	voice                      *voicetranslate.Service
	costCharger                usersCostCharger
	warnErrorsMetric           *metrics.Counter
	useExistingTranslateMetric *metrics.Counter
	translateLatencyMetric     *metrics.Latency
	sendReplyLatencyMetric     *metrics.Latency
}

func New(bot *bot.Bot, voice *voicetranslate.Service, costCharger usersCostCharger) *Handler {
	// Create metrics
	warnErrorsMetric := metrics.NewCounter(
		"bot_voice_handler_warn_errors_count",
		"Counter for warnings in voice message handler",
		[]string{"type"},
	)
	useExistingTranslateMetric := metrics.NewCounter(
		"bot_voice_handler_use_existing_translate_count",
		"Counter for use of existing translation from cache",
		nil,
	)
	translateLatencyMetric := metrics.NewLatency(
		"bot_voice_handler_translate_latency",
		"Latency for the whole voice message translation process",
		nil,
		nil,
	)
	sendReplyLatencyMetric := metrics.NewLatency(
		"bot_voice_handler_send_reply_latency",
		"Latency for the reply with voice message metric",
		nil,
		[]string{"source"},
	)

	return &Handler{
		bot:                        bot,
		voice:                      voice,
		costCharger:                costCharger,
		warnErrorsMetric:           warnErrorsMetric,
		useExistingTranslateMetric: useExistingTranslateMetric,
		translateLatencyMetric:     translateLatencyMetric,
		sendReplyLatencyMetric:     sendReplyLatencyMetric,
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
		totalLatency := h.translateLatencyMetric.NewAction()
		// Send recording voice message status
		wem := h.warnErrorsMetric.NewAction("chat_record_audio")
		if err := h.bot.SendChatAction(update.Message.Chat.ID, tgbotapi.ChatRecordAudio); err != nil {
			log.Warn().
				Err(err).
				Msg("Failed to send chat action")
			wem.Inc(metrics.StatusErr)
		}

		// Ge bytes from voice message
		v := update.Message.Voice
		fileURL, err := h.bot.GetFileDirectURL(v.FileID)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to get file direct URL")
			totalLatency.Observe(metrics.StatusErr)
			return
		}
		data, err := getBytesFromURL(fileURL)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to get bytes from file URL")
			totalLatency.Observe(metrics.StatusErr)
			return
		}

		// Translate voice message
		user := users.Ctx(ctx)
		genSpeech, err := h.voice.Translate(ctx, data, v.Duration, user.SourceLang, user.TargetLang)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to translate voice message")
			totalLatency.Observe(metrics.StatusErr)
			return
		}

		// Charge cost
		wem = h.warnErrorsMetric.NewAction("charge_cost")
		err = h.costCharger.ChargeCost(ctx, user.TelegramUserID, genSpeech.Cost)
		if err != nil {
			log.Warn().
				Err(err).
				Int64("cost", genSpeech.Cost).
				Int64("telegram_user_id", user.TelegramUserID).
				Msg("Failed to charge action cost for user")
			wem.Inc(metrics.StatusErr)
		}

		etm := h.useExistingTranslateMetric.NewAction()
		if !genSpeech.UseExisting() {
			etm.Inc(metrics.StatusErr)
			msg := tgbotapi.NewVoiceUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Bytes: genSpeech.Voice})
			msg.ReplyToMessageID = update.Message.MessageID
			sentMsg, err := h.bot.Send(msg)
			if err != nil {
				log.Error().
					Err(err).
					Msg("error while trying to send reply message")
				totalLatency.Observe(metrics.StatusErr)
				return
			}
			if sentMsg.Voice != nil {
				wem = h.warnErrorsMetric.NewAction("flush_translate_cache")
				if err := genSpeech.FlushCache(ctx, sentMsg.Voice.FileID); err != nil {
					log.Warn().
						Err(err).
						Msg("Failed to flush speech cache")
					wem.Inc(metrics.StatusOk)
				}
			} else {
				wem = h.warnErrorsMetric.NewAction("voice_msg_sent_nil")
				log.Warn().
					Msg("voice in sent message is nil")
				wem.Inc(metrics.StatusErr)
			}
		} else {
			etm.Inc(metrics.StatusOk)
			msg := tgbotapi.NewVoiceShare(update.Message.Chat.ID, genSpeech.FileID)
			if _, err := h.bot.Send(msg); err != nil {
				log.Error().
					Err(err).
					Msg("Failed to send message to user")
				totalLatency.Observe(metrics.StatusErr)
				return
			}
		}
		totalLatency.Observe(metrics.StatusOk)
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
