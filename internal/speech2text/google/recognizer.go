package google

import (
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"errors"
	"fmt"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

type Recognizer struct {
	client *speech.Client
}

func New(client *speech.Client) *Recognizer {
	return &Recognizer{client: client}
}

func (r *Recognizer) ToText(ctx context.Context, data []byte, lang string) (string, error) {
	resp, err := r.client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			AudioChannelCount:          1,
			EnableAutomaticPunctuation: true,
			Encoding:                   speechpb.RecognitionConfig_OGG_OPUS,
			LanguageCode:               lang,
			SampleRateHertz:            48000,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to recognize speech: %w", err)
	}
	if len(resp.Results) < 1 {
		return "", errors.New("google cloud speech API rerturned an empty response")
	}
	return resp.Results[0].Alternatives[0].Transcript, nil
}
