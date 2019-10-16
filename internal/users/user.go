package users

import "context"

type User struct {
	TelegramUserID int64
	UserName       string
	FirstName      string
	LastName       string
	Lang           string
	SourceLang     string
	TargetLang     string
	Points         int64
}

type ctxKey struct{}

func Ctx(ctx context.Context) *User {
	if u, ok := ctx.Value(ctxKey{}).(*User); ok {
		return u
	}
	return nil
}

func (user *User) WithContext(ctx context.Context) context.Context {
	if u, ok := ctx.Value(ctxKey{}).(*User); ok {
		if u == user {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxKey{}, user)
}
