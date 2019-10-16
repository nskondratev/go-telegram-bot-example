package logger

import (
	"io"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func New(l string, w io.Writer) (zerolog.Logger, error) {
	logger := zerolog.New(w).With().Timestamp().Caller().Logger()
	pl, err := zerolog.ParseLevel(l)
	if err != nil {
		return logger, errors.Wrap(err, "failed to parse logger level")
	}

	zerolog.SetGlobalLevel(pl)

	return logger, nil
}
