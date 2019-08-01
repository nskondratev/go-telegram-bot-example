package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Handler interface {
	Handle(ctx context.Context, update tgbotapi.Update)
}

type HandleFunc func(ctx context.Context, update tgbotapi.Update)

func (f HandleFunc) Handle(ctx context.Context, update tgbotapi.Update) {
	f(ctx, update)
}

type Middleware func(next Handler) Handler
