package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Bot struct {
	tg      *tgbotapi.BotAPI
	logger  zerolog.Logger
	handler Handler
}

func New(logger zerolog.Logger, apiToken string) (Bot, error) {
	b := Bot{
		logger: logger.With().Str("module", "bot").Logger(),
	}

	if apiToken == "" {
		return b, errors.New("telegram api token must be provided")
	}

	tg, err := tgbotapi.NewBotAPI(apiToken)

	if err != nil {
		return b, errors.Wrap(err, "failed to create telegram bot instance")
	}

	//if zerolog.GlobalLevel() == zerolog.DebugLevel {
	//	tg.Debug = true
	//}

	b.tg = tg

	return b, nil
}

func (b *Bot) Handle(h Handler) {
	b.handler = h
}

func (b Bot) RunUpdateChannel(ctx context.Context) error {
	if b.handler == nil {
		return errors.New("handler must be set before running updater")
	}

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
				return nil
			}

			handlerLogger := b.logger.With().Str("requestId", uuid.New().String()).Logger()

			handlerCtx := handlerLogger.WithContext(ctx)

			b.handler.Handle(handlerCtx, update)
		}
	}
}

func (b Bot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.tg.Send(c)
}

func (b Bot) GetFileDirectURL(fileID string) (string, error) {
	return b.tg.GetFileDirectURL(fileID)
}

func (b Bot) SendChatAction(chatID int64, action string) error {
	_, err := b.tg.Send(tgbotapi.NewChatAction(chatID, action))
	return err
}
