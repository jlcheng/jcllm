package openai

import (
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
	//TODO implement me
	panic("implement me")
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

var _ llm.ProviderIfc = (*Provider)(nil)
