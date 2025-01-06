package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/go-errors/errors"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"slices"
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

type ModelClient struct {
	config      configuration.Configuration
	client      *genai.Client
	model       string
	chatSession *genai.ChatSession
}

func (g *ModelClient) SolicitResponse(ctx context.Context, conversation llm.Conversation) (llm.ResponseStream, error) {
	client := g.client.GenerativeModel(g.model)
	client.SafetySettings = blockCategoriesNone
	var chatSession = g.chatSession
	if chatSession == nil {
		chatSession = client.StartChat()
		g.chatSession = chatSession
	}
	chatSession.History = []*genai.Content{}
	remoteStream := chatSession.SendMessageStream(ctx, genai.Text(conversation.Entries[len(conversation.Entries)-1].Text))
	exchange := make(chan llm.Message)
	response := llm.ResponseStream{
		Role:           g.ToGenericRole(RoleModel),
		ResponseStream: exchange,
	}
	go func() {
		for {
			respChunk, err := remoteStream.Next()
			if err != nil {
				// End of stream
				if errors.Is(err, iterator.Done) {
					close(exchange)
					break
				}

				// Otherwise, map the error from the genai library into something user-friendly
				exchange <- llm.Message{Err: mapStreamReadError(err)}
				break
			}
			buf := new(bytes.Buffer)
			for _, part := range respChunk.Candidates[0].Content.Parts {
				switch v := part.(type) {
				case genai.Text:
					buf.WriteString(string(v))
				case genai.Blob:
					buf.WriteString(base64.StdEncoding.EncodeToString(v.Data))
				case genai.FunctionResponse:
					buf.WriteString(fmt.Sprintf("(%s)", v.Name))
				}
			}
			exchange <- llm.Message{Text: buf.String()}
		}
	}()
	return response, nil
}

func (g *Gemini) GetModel(_ context.Context, model string) (llm.ModelIfc, error) {
	return &ModelClient{g.config, g.client, model, nil}, nil
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

func (g *ModelClient) ToProviderRole(genericRole string) (providerRole string) {
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

func (g *ModelClient) ToGenericRole(providerRole string) (genericRole string) {
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
	}
	return results, nil
}

func harmBlockNone() []*genai.SafetySetting {
	var harmCategories = []genai.HarmCategory{
		genai.HarmCategoryHateSpeech,
		genai.HarmCategorySexuallyExplicit,
		genai.HarmCategoryDangerousContent,
		genai.HarmCategoryHarassment,
	}
	return slices.Collect(it.Map(slices.Values(harmCategories), func(category genai.HarmCategory) *genai.SafetySetting {
		return &genai.SafetySetting{
			Category:  category,
			Threshold: genai.HarmBlockNone,
		}
	}))
}

var blockCategoriesNone = harmBlockNone()

var _ llm.ProviderIfc = (*Gemini)(nil)
var _ llm.ModelIfc = (*ModelClient)(nil)
var _ llm.RoleMapper = (*ModelClient)(nil)
