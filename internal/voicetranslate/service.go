package voicetranslate

import (
	"context"
	"errors"
	"fmt"
	"github.com/nskondratev/go-telegram-translator-bot/internal/util"
	"math"
	"unicode/utf8"

	"github.com/rs/zerolog"
)

// Service that transforms speech to text
type SpeechToTexter interface {
	ToText(ctx context.Context, data []byte, lang []string) (text, recognizedLang string, err error)
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
	Cost       int64
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
func (s *Service) Translate(ctx context.Context, voice []byte, duration int, sourceLang, targetLang string) (TranslateResult, error) {
	log := zerolog.Ctx(ctx)
	var res TranslateResult
	recognizedSpeech, recognizedLang, err := s.s2t.ToText(ctx, voice, []string{sourceLang, targetLang})
	if err != nil {
		return res, fmt.Errorf("failed to recognize text from speech: %w", err)
	}
	res.Cost += int64(math.Ceil(float64(duration) / 15.0))
	if res.Cost < 1 {
		res.Cost = 1
	}
	targetLang = util.GetTargetLang(recognizedLang, sourceLang, targetLang)
	translated, err := s.tc.Get(ctx, recognizedSpeech, recognizedLang, targetLang)
	if err != nil {
		if !errors.Is(err, ErrNotFoundInCache) {
			log.Warn().
				Err(err).
				Msg("failed to lookup text translation in cache")
		}
		translated, err = s.textTr.Translate(ctx, recognizedSpeech, recognizedLang, targetLang)
		if err != nil {
			return res, fmt.Errorf("failed to translate text: %w", err)
		}
		translateCost := int64(math.Ceil(float64(utf8.RuneCountInString(recognizedSpeech) / 100.0)))
		if translateCost < 1 {
			translateCost = 1
		}
		res.Cost += translateCost
		err := s.tc.Store(ctx, recognizedSpeech, translated, recognizedLang, targetLang)
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
		genSpeechCost := int64(math.Ceil(float64(duration) / 15.0))
		if genSpeechCost < 1 {
			genSpeechCost = 1
		}
		res.Cost += genSpeechCost
		res.SetCacheFlusher(s.sc, translated, targetLang)
	}
	res.Voice = generatedSpeech.Data
	res.FileID = generatedSpeech.FileID
	return res, nil
}
