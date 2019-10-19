package google

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

type Translator struct {
	client *translate.Client
}

func New(client *translate.Client) *Translator {
	return &Translator{client: client}
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
	translations, err := t.client.Translate(ctx, []string{text}, target, &translate.Options{
		Source: source,
		Format: translate.Text,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get translations from google cloud api: %w", err)
	}
	if len(translations) < 1 {
		return "", errors.New("google cloud api returned empty translations response")
	}
	return translations[0].Text, nil
}
