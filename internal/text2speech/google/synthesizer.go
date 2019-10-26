package google

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"golang.org/x/text/language"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"

	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
	"github.com/nskondratev/go-telegram-translator-bot/internal/voicetranslate"
)

type Synthesizer struct {
	client        *texttospeech.Client
	latencyMetric *metrics.Latency
}

func New(client *texttospeech.Client) *Synthesizer {
	return &Synthesizer{
		client: client,
		latencyMetric: metrics.NewLatency(
			"bot_gcloud_text2speech_latency",
			"Latency for speech synthesis requests to GCloud",
			nil,
			nil,
		),
	}
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
	m := s.latencyMetric.NewAction()
	resp, err := s.client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		m.Observe(metrics.StatusErr)
		return res, fmt.Errorf("failed to synthesize speech: %w", err)
	}
	m.Observe(metrics.StatusOk)
	res.Data = resp.AudioContent
	return res, nil
}
