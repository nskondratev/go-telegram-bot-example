package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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

	b, err := bot.NewBot(logger, cnf.Telegram.ApiToken)

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
	signal.Notify(ic, os.Interrupt)

	<-ic
	logger.Info().
		Msg("application is interrupted. Stopping appCtx...")
	appCancel()
	time.Sleep(500 * time.Millisecond)
}
