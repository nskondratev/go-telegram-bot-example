package google

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"golang.org/x/text/language"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"

	"github.com/nskondratev/go-telegram-translator-bot/internal/voicetranslate"
)

type Synthesizer struct {
	client *texttospeech.Client
}

func New(client *texttospeech.Client) *Synthesizer {
	return &Synthesizer{client: client}
}

func (s *Synthesizer) ToSpeech(ctx context.Context, text string, lang string) (voicetranslate.TextToSpeechResult, error) {
	var res voicetranslate.TextToSpeechResult
	target, err := language.Parse(lang)
	if err != nil {
		return res, fmt.Errorf("failed to parse language: %w", err)
	}
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: target.String(),
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   texttospeechpb.AudioEncoding_OGG_OPUS,
			SampleRateHertz: 48000,
		},
	}
	resp, err := s.client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return res, fmt.Errorf("failed to synthesize speech: %w", err)
	}
	res.Data = resp.AudioContent
	return res, nil
}
