package google

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"

	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
)

type Translator struct {
	client        *translate.Client
	latencyMetric *metrics.Latency
}

func New(client *translate.Client) *Translator {
	return &Translator{
		client: client,
		latencyMetric: metrics.NewLatency(
			"bot_gcloud_text_translate_latency",
			"Latency for translation requests to GCloud",
			nil,
			nil,
		),
	}
}

func (t *Translator) Translate(ctx context.Context, text string, sourceLang, targetLang string) (string, error) {
	target, err := language.Parse(targetLang)
	if err != nil {
		return "", fmt.Errorf("failed to parse target language: %w", err)
	}
	source, err := language.Parse(sourceLang)
	if err != nil {
		return "", fmt.Errorf("failed to parse source language: %w", err)
	}
	m := t.latencyMetric.NewAction()
	translations, err := t.client.Translate(ctx, []string{text}, target, &translate.Options{
		Source: source,
		Format: translate.Text,
	})
	if err != nil {
		m.Observe(metrics.StatusErr)
		return "", fmt.Errorf("failed to get translations from google cloud api: %w", err)
	}
	if len(translations) < 1 {
		m.Observe(metrics.StatusErr)
		return "", errors.New("google cloud api returned empty translations response")
	}
	m.Observe(metrics.StatusOk)
	return translations[0].Text, nil
}
