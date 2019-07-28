package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Bot struct {
	tg     *tgbotapi.BotAPI
	logger zerolog.Logger
}

func NewBot(logger zerolog.Logger, apiToken string) (Bot, error) {
	b := Bot{logger: logger.With().Str("module", "bot").Logger()}

	if apiToken == "" {
		return b, errors.New("telegram api token must be provided")
	}

	tg, err := tgbotapi.NewBotAPI(apiToken)

	if err != nil {
		return b, errors.Wrap(err, "failed to create telegram bot instance")
	}

	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		tg.Debug = true
	}

	b.tg = tg

	return b, nil
}

func (b Bot) RunUpdateChannel(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := b.tg.GetUpdatesChan(updateConfig)

	if err != nil {
		return errors.Wrap(err, "failed to get updates channel")
	}

	for {
		select {
		case <-ctx.Done():
			b.logger.Info().
				Str("cause", "context is closed").
				Msg("exit loop for getting updates")
			return nil
		case update, ok := <-updates:
			if !ok {
				b.logger.Info().
					Str("cause", "updates channel is closed").
					Msg("exit loop for getting updates")
			}

			if update.Message == nil {
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			msg.ReplyToMessageID = update.Message.MessageID

			if _, err := b.tg.Send(msg); err != nil {
				b.logger.Error().
					Err(err).
					Msg("failed to send message")
			}
		}
	}
}
