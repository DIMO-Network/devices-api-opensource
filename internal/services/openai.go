package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type OpenAI interface {
	GetErrorCodesDescription(make, model string, errorCodes []string) (string, error)
}

type openAI struct {
	chatGptURL string
	token      string
	httpClient *http.Client
	logger     *zerolog.Logger
}

type ChatGPTResponseChoices struct {
	Message struct {
		Role    string
		Content string
	}
	Index        int
	FinishReason string `json:"finish_reason"`
}

type ChatGPTResponse struct {
	ID      string
	Object  string
	Created int
	Model   string
	Usage   struct {
		PromptTokens     int
		CompletionTokens int
		TotalTokens      int
	}
	Choices []ChatGPTResponseChoices
}

func NewOpenAI(logger *zerolog.Logger, c config.Settings) OpenAI {
	return &openAI{
		chatGptURL: c.ChatGPTURL,
		token:      c.OpenAISecretKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (o *openAI) askChatGPT(body io.Reader) (*ChatGPTResponse, error) {
	req, err := http.NewRequest(
		"POST",
		o.chatGptURL,
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.token)

	var currentReqResponseTime time.Duration
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			currentReqResponseTime = time.Since(time.Now())
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	appmetrics.OpenAIResponseTimeOps.With(prometheus.Labels{
		"api_url": o.chatGptURL,
		"status":  strconv.Itoa(resp.StatusCode),
	}).Observe(currentReqResponseTime.Seconds())
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received error from request: %s", string(b))
	}

	cResp := &ChatGPTResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response json: %w", err)
	}

	return cResp, nil
}

func (o *openAI) GetErrorCodesDescription(make, model string, errorCodes []string) (string, error) {
	codes := strings.Join(errorCodes, ", ")
	req := fmt.Sprintf(`{
		"model": "gpt-3.5-turbo",
		"messages": [
			{
				"role": "user", 
				"content": "Briefly summarize for me, in order, what the following error codes mean for a %s %s. The error codes are %s."}]
	  		}
	  `, make, model, codes)

	r, err := o.askChatGPT(strings.NewReader(req))
	if err != nil {
		return "", err
	}

	appmetrics.OpenAITotalTokensUsedOps.Add(float64(r.Usage.TotalTokens))

	if len(r.Choices) == 0 {
		return "", errors.New("could not fetch description for error codes")
	}

	c := r.Choices[0]
	if c.FinishReason != "stop" {
		o.logger.Error().Interface("rawResponse", r).Msg("Unexpected finish_reason from ChatGPT.")
	}

	return strings.Trim(c.Message.Content, "\n"), nil
}
