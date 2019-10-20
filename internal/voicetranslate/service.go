package voicetranslate

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
)

// Service that transforms speech to text
type SpeechToTexter interface {
	ToText(ctx context.Context, data []byte, lang string) (string, error)
}

// Service that translates text from one language to another
type TextTranslator interface {
	Translate(ctx context.Context, text string, sourceLang, targetLang string) (string, error)
}

type TextToSpeechResult struct {
	Data   []byte
	FileID string
}

func (t TextToSpeechResult) UseExisting() bool {
	return len(t.FileID) > 0
}

// Service that transforms text to speech
type TextToSpeecher interface {
	ToSpeech(ctx context.Context, text string, lang string) (TextToSpeechResult, error)
}

type TranslateResult struct {
	Voice      []byte
	FileID     string
	FlushCache SpeechCacheStorer
}

func (tr TranslateResult) UseExisting() bool {
	return len(tr.FileID) > 0
}

func (tr *TranslateResult) SetCacheFlusher(sc SpeechCache, text, lang string) {
	tr.FlushCache = func(ctx context.Context, fileID string) error {
		return sc.Store(ctx, fileID, text, lang)
	}
}

// Performs the translation of voice message
type Service struct {
	s2t    SpeechToTexter
	textTr TextTranslator
	t2s    TextToSpeecher
	tc     TranslatorCache
	sc     SpeechCache
}

// Create new voice translation service
func New(
	s2t SpeechToTexter,
	textTr TextTranslator,
	t2s TextToSpeecher,
	tc TranslatorCache,
	sc SpeechCache,
) *Service {
	return &Service{
		s2t:    s2t,
		textTr: textTr,
		t2s:    t2s,
		tc:     tc,
		sc:     sc,
	}
}

// Perform voice translation
func (s *Service) Translate(ctx context.Context, voice []byte, sourceLang, targetLang string) (TranslateResult, error) {
	log := zerolog.Ctx(ctx)
	var res TranslateResult
	recognizedSpeech, err := s.s2t.ToText(ctx, voice, sourceLang)
	if err != nil {
		return res, fmt.Errorf("failed to recognize text from speech: %w", err)
	}
	translated, err := s.tc.Get(ctx, recognizedSpeech, sourceLang, targetLang)
	if err != nil {
		if !errors.Is(err, ErrNotFoundInCache) {
			log.Warn().
				Err(err).
				Msg("failed to lookup text translation in cache")
		}
		translated, err = s.textTr.Translate(ctx, recognizedSpeech, sourceLang, targetLang)
		if err != nil {
			return res, fmt.Errorf("failed to translate text: %w", err)
		}
		err := s.tc.Store(ctx, recognizedSpeech, translated, sourceLang, targetLang)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("failed to lookup text translation in cache")
		}
	}
	generatedSpeech, err := s.sc.Get(ctx, translated, targetLang)
	if err != nil {
		if !errors.Is(err, ErrNotFoundInCache) {
			log.Warn().
				Err(err).
				Msg("failed to lookup generated speech in cache")
		}
		generatedSpeech, err = s.t2s.ToSpeech(ctx, translated, targetLang)
		if err != nil {
			return res, fmt.Errorf("failed to generate speech: %w", err)
		}
		res.SetCacheFlusher(s.sc, translated, targetLang)
	}
	res.Voice = generatedSpeech.Data
	res.FileID = generatedSpeech.FileID
	return res, nil
}
