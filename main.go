package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nskondratev/go-telegram-bot-example/bot"
	"github.com/nskondratev/go-telegram-bot-example/conf"
	appLogger "github.com/nskondratev/go-telegram-bot-example/log"
)

func main() {
	cnf, err := conf.NewConf()

	if err != nil {
		log.Fatal(err)
	}

	logger, err := appLogger.NewLogger(cnf.LogLevel, os.Stdout)

	if err != nil {
		log.Fatal(err)
	}

	// Construct update handlers

	logTimeExecution := func(next bot.Handler) bot.Handler {
		return bot.HandleFunc(func(ctx context.Context, update tgbotapi.Update) {
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

	logUsername := func(next bot.Handler) bot.Handler {
		return bot.HandleFunc(func(ctx context.Context, update tgbotapi.Update) {
			username := "unknown"
			if update.Message != nil && update.Message.From != nil {
				username = update.Message.From.UserName
			}
			zerolog.Ctx(ctx).Info().
				Str("username", username).
				Int("updateID", update.UpdateID).
				Msg("Log username and update id")
			next.Handle(ctx, update)
		})
	}

	b, err := bot.NewBot(logger, cnf.Telegram.ApiToken)

	handler := bot.
		NewChain(
			logTimeExecution,
			logUsername,
		).
		ThenFunc(func(ctx context.Context, update tgbotapi.Update) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			if _, err := b.Send(msg); err != nil {
				zerolog.Ctx(ctx).Error().
					Err(err).
					Msg("error while trying to send reply message")
			}
		})

	b.Handle(handler)

	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to create bot")
	}

	appCtx, appCancel := context.WithCancel(context.Background())

	go func() {
		err := b.RunUpdateChannel(appCtx)

		if err != nil {
			logger.Fatal().
				Err(err).
				Msg("error in bot update channel listener")
		}
	}()

	logger.Info().
		Str("telegram api token", os.Getenv(cnf.Telegram.ApiToken)).
		Msg("Start telegram bot application")

	// Wait for interruption
	ic := make(chan os.Signal, 1)
	signal.Notify(ic, os.Interrupt, syscall.SIGTERM)

	<-ic
	logger.Info().
		Msg("application is interrupted. Stopping appCtx...")
	appCancel()
	time.Sleep(500 * time.Millisecond)
}
