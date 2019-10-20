package voicetranslate

import (
	"context"
	"errors"
)

var ErrNotFoundInCache = errors.New("not found in cache")

type TranslatorCache interface {
	Get(ctx context.Context, text, sourceLang, targetLang string) (string, error)
	Store(ctx context.Context, text, translation, sourceLang, targetLang string) error
}

type SpeechCache interface {
	Get(ctx context.Context, text, lang string) (TextToSpeechResult, error)
	Store(ctx context.Context, fileID, text, lang string) error
}

type SpeechCacheStorer func(ctx context.Context, fileID string) error
