package middleware

import (
	"context"
	"errors"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"

	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/users"
)

func LogTimeExecution(next bot.Handler) bot.Handler {
	return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
		logger := zerolog.Ctx(ctx)
		logger.Info().Msg("Set up log time execution middleware")
		ts := time.Now()
		next.Handle(ctx, update)
		te := time.Now().Sub(ts)
		logger.Info().
			Str("executionTime", te.String()).
			Msg("Log time execution")
	})
}

func LogUserInfo(next bot.Handler) bot.Handler {
	return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
		username := "unknown"
		if update.Message != nil && update.Message.From != nil {
			username = update.Message.From.UserName
		}
		zerolog.Ctx(ctx).Info().
			Str("username", username).
			Int("updateID", update.UpdateID).
			Msg("New update from user")
		next.Handle(ctx, update)
	})
}

func SetUser(usersStore users.Store) func(next bot.Handler) bot.Handler {
	return func(next bot.Handler) bot.Handler {
		return bot.HandlerFunc(func(ctx context.Context, update tgbotapi.Update) {
			logger := zerolog.Ctx(ctx)
			if update.Message != nil && update.Message.From != nil {
				user, err := usersStore.GetUserByTelegramUserID(ctx, int64(update.Message.From.ID))
				if err != nil {
					logger.Error().
						Err(err).
						Int64("tg_user_id", int64(update.Message.From.ID)).
						Msg("User not found")
					if errors.Is(err, users.ErrUserNotFound) {
						user = users.User{
							TelegramUserID: int64(update.Message.From.ID),
							UserName:       update.Message.From.UserName,
							FirstName:      update.Message.From.FirstName,
							LastName:       update.Message.From.LastName,
							Lang:           update.Message.From.LanguageCode,
							SourceLang:     "ru",
							TargetLang:     "en",
							Points:         60,
						}
						if err := usersStore.StoreUser(ctx, &user); err != nil {
							logger.Error().
								Err(err).
								Msg("failed to store user in store")
						}
					} else {
						logger.Error().
							Err(err).
							Msg("failed to fetch user by tg userID from store")
					}
				}
				ctx = user.WithContext(ctx)
			} else {
				logger.Info().Msg("can not get sender user from this update")
			}
			next.Handle(ctx, update)
		})
	}
}
