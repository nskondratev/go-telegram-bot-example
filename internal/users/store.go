package users

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type Store interface {
	GetUserByTelegramUserID(ctx context.Context, tgUserID int64) (User, error)
	StoreUser(ctx context.Context, user *User) error
	UpdateTranslationLangs(ctx context.Context, tgUserID int64, sourceLang, targetLang string) error
}
