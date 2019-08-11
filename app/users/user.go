package users

type User struct {
	TelegramUserID int64   `bson:"telegramUserID"`
	UserName       string  `bson:"userName"`
	FirstName      string  `bson:"firstName"`
	LastName       string  `bson:"lastName"`
	UserLang       string  `bson:"userLang"`
	SourceLang     string  `bson:"sourceLang"`
	TargetLang     string  `bson:"targetLang"`
	Points         float64 `bson:"points"`
}
