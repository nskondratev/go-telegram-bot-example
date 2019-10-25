package google

import (
	speech "cloud.google.com/go/speech/apiv1p1beta1"
	"context"
	"errors"
	"fmt"
	"github.com/nskondratev/go-telegram-translator-bot/internal/util"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
)

type Recognizer struct {
	client *speech.Client
}

func New(client *speech.Client) *Recognizer {
	return &Recognizer{client: client}
}

func (r *Recognizer) ToText(ctx context.Context, data []byte, lang []string) (string, string, error) {
	resp, err := r.client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			AudioChannelCount:          1,
			EnableAutomaticPunctuation: true,
			Encoding:                   speechpb.RecognitionConfig_OGG_OPUS,
			LanguageCode:               lang[0],
			SampleRateHertz:            16000,
			AlternativeLanguageCodes:   lang[1:],
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to recognize speech: %w", err)
	}
	if len(resp.Results) < 1 {
		return "", "", errors.New("google cloud speech API rerturned an empty response")
	}
	text, recognized := getBestResult(resp.Results)
	return text, recognized, nil
}

// Get best result by confidence
// Returns text and language code
func getBestResult(results []*speechpb.SpeechRecognitionResult) (text string, lang string) {
	var conf float32
	for _, res := range results {
		for _, a := range res.Alternatives {
			if a.Confidence > conf {
				conf = a.Confidence
				text = a.Transcript
				lang = res.LanguageCode
			}
		}
	}
	return text, util.Normalize(lang)
}
