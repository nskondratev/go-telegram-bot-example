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
	UpdateSourceLang(ctx context.Context, tgUserID int64, sourceLang string) error
	UpdateTargetLang(ctx context.Context, tgUserID int64, targetLang string) error
	ChargeCost(ctx context.Context, tgUserID, cost int64) error
}
