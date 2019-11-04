package lang

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

type Language struct {
	Code  string
	Label string
}

var (
	English = Language{Code: "en", Label: "English 🇬🇧"}
	Russian = Language{Code: "ru", Label: "Russian 🇷🇺"}
	Spanish = Language{Code: "es", Label: "Spanish 🇪🇸"}
	French  = Language{Code: "fr", Label: "French 🇫🇷"}
	Deutsch = Language{Code: "de", Label: "Deutsch 🇩🇪"}
)

func buildKeyboard(dataPrefix string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(Russian.Label, dataPrefix+":"+Russian.Code),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(English.Label, dataPrefix+":"+English.Code),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(Spanish.Label, dataPrefix+":"+Spanish.Code),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(French.Label, dataPrefix+":"+French.Code),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(Deutsch.Label, dataPrefix+":"+Deutsch.Code),
		),
	)
}

func Normalize(langCode string) (res string) {
	tokens := strings.Split(langCode, "-")
	if len(tokens) > 0 {
		res = tokens[0]
	}
	return res
}

func GetTargetLang(recognized, source, target string) string {
	norm := Normalize(recognized)
	if norm == target {
		return source
	} else {
		return target
	}
}
