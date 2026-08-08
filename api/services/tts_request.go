package services

import (
	"context"
	"io"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ttsResponse struct {
	Body      io.ReadCloser
	RequestID string
}

func requestOpenAITTS(ctx context.Context, apiKey, baseURL, model, voice, text string) (*ttsResponse, error) {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if strings.TrimSpace(baseURL) != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	client := openai.NewClient(opts...)
	res, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.SpeechModel(model),
		Input: text,
		Voice: openai.AudioSpeechNewParamsVoiceUnion{
			OfAudioSpeechNewsVoiceString2: openai.String(voice),
		},
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
	})
	if err != nil {
		return nil, err
	}

	requestID := strings.TrimSpace(res.Header.Get("x-request-id"))
	if requestID == "" {
		requestID = strings.TrimSpace(res.Header.Get("request-id"))
	}
	return &ttsResponse{Body: res.Body, RequestID: requestID}, nil
}
