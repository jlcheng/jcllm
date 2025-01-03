package genaigo

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"strings"
)

const (
	// RoleUser is 'user'
	RoleUser = "user"
	// RoleModel is  'model', as Gemini only recognize 'user' or 'model'.
	RoleModel = "model"
)

// Gemini implements a llm.ModelIfc using the Vertex AI Go SDK
type Gemini struct {
	config configuration.Configuration
	client *genai.Client
}

func (g *Gemini) GetModel(ctx context.Context, modelName string) (llm.ModelIfc, error) {
	return nil, llm.ErrNoSupport
}

func New(ctx context.Context, config configuration.Configuration) (*Gemini, error) {
	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(
		config.String("gemini-api-key"),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize generative-ai-go/genai client: %w", err)
	}

	return &Gemini{
		config: config,
		client: genaiClient,
	}, nil
}

func (g *Gemini) ToProviderRole(genericRole string) (providerRole string) {
	switch genericRole {
	case llm.RoleAssistant:
		return RoleModel
	case llm.RoleUser:
		return RoleUser
	case llm.RolesSystem:
		return RoleUser
	}
	return genericRole

}

func (g *Gemini) ToGenericRole(providerRole string) (genericRole string) {
	switch providerRole {
	case RoleModel:
		return llm.RoleAssistant
	case RoleUser:
		return llm.RoleUser
	}
	return genericRole
}

func (g *Gemini) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	results := make([]llm.ModelInfo, 0)
	modelsIterator := g.client.ListModels(ctx)
	for {
		elem, err := modelsIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, errors.New(err)
		}
		if !strings.Contains(strings.ToLower(elem.DisplayName), "gemini") {
			continue
		}
		results = append(results, llm.ModelInfo{
			Name:        strings.TrimPrefix(elem.Name, "models/"),
			DisplayName: elem.DisplayName,
			Description: elem.DisplayName + ": " + elem.Description,
			MaxTokens:   int(elem.InputTokenLimit),
			Version:     elem.Version,
		})

		fmt.Println(elem.Version)
		fmt.Printf("%+v\n", elem.SupportedGenerationMethods)
	}
	return results, nil
}

func (g *Gemini) SolicitResponse(context.Context, llm.Conversation) (llm.ResponseStream, error) {
	return llm.ResponseStream{}, llm.ErrNoSupport
}

var _ llm.ProviderIfc = (*Gemini)(nil)
var _ llm.ModelIfc = (*Gemini)(nil)
var _ llm.RoleMapper = (*Gemini)(nil)
