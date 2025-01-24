package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/go-errors/errors"
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
	"github.com/jlcheng/jcllm/llm"
	"github.com/jlcheng/jcllm/llm/openaimodels"
)

const (
	// RoleUser is 'user'
	RoleUser = "user"
	// RoleAssistant is 'assistant'
	RoleAssistant = "assistant"
	// RoleDeveloper is 'developer
	RoleDeveloper = "developer"

	// HeaderAuthorization is where OpenAI looks for the OpenAI API Key
	HeaderAuthorization = "Authorization"
)

type Provider struct {
	config     configuration.Configuration
	httpClient *http.Client
}

func NewProvider(config configuration.Configuration) *Provider {
	timeout := time.Duration(config.MustInt(keys.OptionHttpTimeout)) * time.Second
	return &Provider{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *Provider) ToProviderRole(genericRole string) (providerRole string) {
	switch genericRole {
	case llm.RoleAssistant:
		return RoleAssistant
	case llm.RoleUser:
		return RoleUser
	case llm.RoleSystem:
		return RoleUser
	}
	return RoleUser
}

func (p *Provider) ToGenericRole(providerRole string) (genericRole string) {
	switch providerRole {
	case RoleAssistant:
		return llm.RoleAssistant
	case RoleUser:
		return llm.RoleUser
	}
	return llm.RoleUser

}

func (p *Provider) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpointURL("/models"), nil)
	if err != nil {
		return nil, errors.WrapPrefix(err, "list model request creation failed", 0)
	}
	body, err := p.submitRequest(request)
	if err != nil {
		return nil, errors.WrapPrefix(err, "list models request submission failed", 0)
	}
	var models openaimodels.ListModelsResponse
	if err := json.NewDecoder(body).Decode(&models); err != nil {
		return nil, errors.WrapPrefix(err, "list model response read failed", 0)
	}
	return slices.SortedFunc(
		it.Map(slices.Values(models.Data), func(model openaimodels.Model) llm.ModelInfo {
			return llm.ModelInfo{
				DisplayName: model.ID,
				Name:        model.ID,
				Description: model.ID,
			}
		}),
		func(a llm.ModelInfo, b llm.ModelInfo) int {
			return strings.Compare(a.Name, b.Name)
		},
	), nil
}

func (p *Provider) SolicitResponse(ctx context.Context, input llm.SolicitResponseInput) (llm.ResponseStream, error) {
	exchange := make(chan llm.Message)
	response := llm.ResponseStream{
		Role:           RoleAssistant,
		ResponseStream: exchange,
	}
	messages := slices.Collect(it.Map(slices.Values(input.Conversation.Entries), func(v llm.ChatEntry) openaimodels.Message {
		return openaimodels.Message{
			Content: v.Text,
			Role:    p.ToProviderRole(v.Role),
		}
	}))
	if developerMessage := p.config.String(keys.OptionSystemPrompt); developerMessage != "" {
		messages = append([]openaimodels.Message{{
			Content: developerMessage,
			Role:    RoleDeveloper,
		}}, messages...)
	}
	chatCompletionRequest := openaimodels.CreateChatCompletionRequest{
		Model:    input.ModelName,
		Messages: messages,
		Stream:   ptr(true),
		StreamOptions: &openaimodels.StreamOptions{
			IncludeUsage: ptr(true),
		},
	}
	requestBytes, err := json.Marshal(chatCompletionRequest)
	if err != nil {
		return response, errors.WrapPrefix(err, "chat completion request stringify failed", 0)
	}
	requestAsStream := bytes.NewReader(requestBytes)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpointURL("/chat/completions"), requestAsStream)
	if err != nil {
		return response, errors.WrapPrefix(err, "chat completions request creation failed", 0)
	}
	request.Header.Set("Content-Type", "application/json")
	body, err := p.submitRequest(request)
	if err != nil {
		return response, errors.WrapPrefix(err, "chat completions request submission failed", 0)
	}
	go func() {
		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			lineMessage := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if lineMessage == "[DONE]" {
				break
			}
			var chunk openaimodels.ChatCompletionChunkResponse
			if err := json.Unmarshal([]byte(lineMessage), &chunk); err != nil {
				exchange <- llm.Message{
					Err: errors.WrapPrefix(err, "unmarshal response failed", 0),
				}
				break
			}
			if len(chunk.Choices) != 0 {
				exchange <- llm.Message{
					Text: chunk.Choices[0].Delta.Content,
					Err:  nil,
				}
			}
			if chunk.Usage != nil {
				exchange <- llm.Message{
					TokenCount: chunk.Usage.CompletionTokens,
				}
			}
		}
		close(exchange)
	}()
	return response, nil
}

func (p *Provider) baseURL() string {
	return p.config.String(keys.OptionOpenAIBaseURL)
}

func (p *Provider) endpointURL(suffix string) string {
	base, err := url.Parse(p.baseURL())
	if err != nil {
		return ""
	}
	return base.JoinPath(suffix).String()
}

func (p *Provider) submitRequest(request *http.Request) (io.ReadCloser, error) {
	var (
		response *http.Response
		body     []byte
		err      error
	)
	if request.Header.Get(HeaderAuthorization) == "" {
		request.Header.Set("Authorization", "Bearer "+p.config.String(keys.OptionOpenAIApiKey))
	}
	response, err = p.httpClient.Do(request)
	if err != nil {
		return nil, errors.WrapPrefix(err, "submit request failed", 0)
	}
	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		body, err = io.ReadAll(response.Body)
		if err != nil {
			body = []byte("cannot read response body")
		}
		var errResponse openaimodels.APIErrorResponse
		_ = json.Unmarshal(body, &errResponse)
		errorMessage := errResponse.Error.Message
		if errorMessage == "" {
			errorMessage = fmt.Sprintf("%q", body)
		}
		return nil, errors.Errorf("submit request failed, status code: %d, message: %s", response.StatusCode, errorMessage)
	}
	return response.Body, nil
}

func ptr[T any](obj T) *T {
	return &obj
}

var _ llm.ProviderIfc = (*Provider)(nil)
