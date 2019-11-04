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
		if sender := getSenderFromUpdate(update); sender != nil {
			username = sender.UserName
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
			log := zerolog.Ctx(ctx)
			sender := getSenderFromUpdate(update)
			if sender != nil {
				user, err := usersStore.GetUserByTelegramUserID(ctx, int64(sender.ID))
				if err != nil {
					log.Error().
						Err(err).
						Int64("tg_user_id", int64(sender.ID)).
						Msg("User not found")
					if errors.Is(err, users.ErrUserNotFound) {
						user = users.User{
							TelegramUserID: int64(sender.ID),
							UserName:       sender.UserName,
							FirstName:      sender.FirstName,
							LastName:       sender.LastName,
							Lang:           sender.LanguageCode,
							SourceLang:     "ru",
							TargetLang:     "en",
							Points:         60,
						}
						if err := usersStore.StoreUser(ctx, &user); err != nil {
							log.Error().
								Err(err).
								Msg("failed to store user in store")
						}
					} else {
						log.Error().
							Err(err).
							Msg("failed to fetch user by tg userID from store")
					}
				}
				ctx = user.WithContext(ctx)
			} else {
				log.Info().Msg("can not get sender user from this update")
			}
			next.Handle(ctx, update)
		})
	}
}

func getSenderFromUpdate(update tgbotapi.Update) *tgbotapi.User {
	switch {
	case update.Message != nil && update.Message.From != nil:
		return update.Message.From
	case update.CallbackQuery != nil && update.CallbackQuery.From != nil:
		return update.CallbackQuery.From
	default:
		return nil
	}
}
