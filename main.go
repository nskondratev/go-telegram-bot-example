package main

import (
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	log.Printf("Provided API token: %s", os.Getenv("TELEGRAM_APITOKEN"))
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Fatal(err) // You should add better error handling than this!
	}

	bot.Debug = true // Has the library display every request and response.

	// Create a new UpdateConfig struct with an offset of 0.
	// Future requests can pass a higher offset to ensure there aren't duplicates.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we want to keep the connection open longer and wait for incoming updates.
	// This reduces the number of requests that are made while improving response time.
	updateConfig.Timeout = 60

	// Now we can pass our UpdateConfig struct to the library to start getting updates.
	// The GetUpdatesChan method is opinionated and as such, it is reasonable to implement
	// your own version of it. It is easier to use if you have no special requirements though.
	updates, err := bot.GetUpdatesChan(updateConfig)

	if err != nil {
		log.Fatalf("Error while trying to get updates chan: %s", err.Error())
	}

	// Now we're ready to start going through the updates we're given.
	// Because we have a channel, we can range over it.
	for update := range updates {
		// There are many types of updates. We only care about messages right now,
		// so we should ignore any other kinds.
		if update.Message == nil {
			continue
		}

		// Because we have to create structs for every kind of request,
		// there's a number of helper functions to make creating common
		// types easier. Here, we're using the NewMessage helper which
		// returns a MessageConfig struct.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		// As there's too many fields for each Config to specify in a single
		// function call, we need to modify the result the helper gave us.
		msg.ReplyToMessageID = update.Message.MessageID

		// We're ready to send our message!
		// The Send method is for Configs that return a Message struct.
		// Sending Messages (among many other types) return a Message.
		// In this case, we don't care about the returned Message.
		// We only need to make sure our message went through successfully.
		if _, err := bot.Send(msg); err != nil {
			panic(err) // Again, this is a bad way to handle errors.
		}
	}
}
