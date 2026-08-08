package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ttsResponse struct {
	Body      io.ReadCloser
	RequestID string
}

const defaultFishAudioBaseURL = "https://api.fish.audio"

type fishAudioTTSRequest struct {
	Text                      string  `json:"text"`
	ReferenceID               string  `json:"reference_id"`
	Temperature               float64 `json:"temperature"`
	TopP                      float64 `json:"top_p"`
	ChunkLength               int     `json:"chunk_length"`
	Normalize                 bool    `json:"normalize"`
	Latency                   string  `json:"latency"`
	RepetitionPenalty         float64 `json:"repetition_penalty"`
	ConditionOnPreviousChunks bool    `json:"condition_on_previous_chunks"`
	Format                    string  `json:"format"`
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

func requestFishAudioTTS(ctx context.Context, apiKey, baseURL, model, referenceID, text string) (*ttsResponse, error) {
	payload, err := json.Marshal(fishAudioTTSRequest{
		Text:                      text,
		ReferenceID:               referenceID,
		Temperature:               0,
		TopP:                      0.1,
		ChunkLength:               300,
		Normalize:                 true,
		Latency:                   "normal",
		RepetitionPenalty:         1.2,
		ConditionOnPreviousChunks: true,
		Format:                    "mp3",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fishAudioTTSEndpoint(baseURL), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")
	if model = strings.TrimSpace(model); model != "" {
		req.Header.Set("model", model)
	}

	response, err := (&http.Client{Timeout: time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 8*1024))
		return nil, fmt.Errorf("Fish Audio TTS request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	requestID := strings.TrimSpace(response.Header.Get("x-request-id"))
	if requestID == "" {
		requestID = strings.TrimSpace(response.Header.Get("request-id"))
	}
	return &ttsResponse{Body: response.Body, RequestID: requestID}, nil
}

func fishAudioTTSEndpoint(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultFishAudioBaseURL
	}
	if strings.HasSuffix(baseURL, "/v1/tts") {
		return baseURL
	}
	return baseURL + "/v1/tts"
}
