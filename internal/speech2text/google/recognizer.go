package google

import (
	speech "cloud.google.com/go/speech/apiv1p1beta1"
	"context"
	"errors"
	"fmt"
	"github.com/nskondratev/go-telegram-translator-bot/internal/lang"
	"github.com/nskondratev/go-telegram-translator-bot/internal/metrics"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
)

var sampleRateHertzs = []int32{16000, 48000}

type Recognizer struct {
	client        *speech.Client
	latencyMetric *metrics.Latency
}

func New(client *speech.Client) *Recognizer {
	return &Recognizer{
		client: client,
		latencyMetric: metrics.NewLatency(
			"bot_gcloud_speech2text_latency",
			"Latency for speech2text requests to GCloud",
			nil,
			[]string{"sample_rate_hertz"},
		),
	}
}

func (r *Recognizer) ToText(ctx context.Context, data []byte, lang []string) (text string, recognizedLang string, err error) {
	for _, sampleRateHertz := range sampleRateHertzs {
		m := r.latencyMetric.NewAction(getLabelBySampleRateHertz(sampleRateHertz))
		resp, err := r.client.Recognize(ctx, &speechpb.RecognizeRequest{
			Config: &speechpb.RecognitionConfig{
				AudioChannelCount:          1,
				EnableAutomaticPunctuation: true,
				Encoding:                   speechpb.RecognitionConfig_OGG_OPUS,
				LanguageCode:               lang[0],
				SampleRateHertz:            sampleRateHertz,
				AlternativeLanguageCodes:   lang[1:],
			},
			Audio: &speechpb.RecognitionAudio{
				AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
			},
		})
		if err != nil {
			m.Observe(metrics.StatusErr)
			return "", "", fmt.Errorf("failed to recognize speech: %w", err)
		}
		if len(resp.Results) < 1 {
			m.Observe(metrics.StatusErr)
			continue
		}
		m.Observe(metrics.StatusOk)
		text, recognizedLang = getBestResult(resp.Results)
		return text, recognizedLang, nil
	}
	return "", "", errors.New("google cloud speech API rerturned an empty response")
}

// Get best result by confidence
// Returns text and language code
func getBestResult(results []*speechpb.SpeechRecognitionResult) (text string, sourceLang string) {
	var conf float32
	for _, res := range results {
		for _, a := range res.Alternatives {
			if a.Confidence > conf {
				conf = a.Confidence
				text = a.Transcript
				sourceLang = res.LanguageCode
			}
		}
	}
	return text, lang.Normalize(sourceLang)
}

func getLabelBySampleRateHertz(in int32) string {
	switch in {
	case 16000:
		return "16000"
	case 48000:
		return "48000"
	default:
		return ""
	}
}
