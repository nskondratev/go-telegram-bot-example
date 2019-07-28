package conf

import (
	"flag"
)

type Conf struct {
	LogLevel string
	Telegram struct {
		ApiToken string
	}
}

func NewConf() (Conf, error) {
	c := Conf{}

	// Add variables
	flag.StringVar(&c.LogLevel, "log-level", "info", "Log level")
	flag.StringVar(&c.Telegram.ApiToken, "api-token", "", "Telegram API Token")

	flag.Parse()

	return c, nil
}
