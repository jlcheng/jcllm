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
	"jcheng.org/jcllm/configuration/keys"
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

func NewProvider(ctx context.Context, config configuration.Configuration) (*Gemini, error) {
	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(
		config.String(keys.OptionGeminiApiKey),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize generative-ai-go/genai client: %w", err)
	}

	return &Gemini{
		config: config,
		client: genaiClient,
	}, nil
}

type ModelClient struct {
	config            configuration.Configuration
	client            *genai.Client
	model             string
	cachedChatSession *genai.ChatSession
	cachedModel       *genai.GenerativeModel
}

func (g *ModelClient) SolicitResponse(ctx context.Context, conversation llm.Conversation) (llm.ResponseStream, error) {
	var chatSession = g.chatSession()
	history, lastEntry := extractLast(conversation)
	chatSession.History = slices.Collect(it.Map(slices.Values(history), g.toContent))
	remoteStream := chatSession.SendMessageStream(ctx, g.toContent(lastEntry).Parts...)
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
			if len(respChunk.Candidates) == 0 {
				continue
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
			exchange <- llm.Message{Text: buf.String(), TokenCount: int(respChunk.UsageMetadata.CandidatesTokenCount)}
		}
	}()
	return response, nil
}

func (g *Gemini) GetModel(_ context.Context, model string) (llm.ModelIfc, error) {
	return &ModelClient{config: g.config, client: g.client, model: model, cachedChatSession: nil, cachedModel: nil}, nil
}

func (g *ModelClient) ModelName() string {
	return g.model
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

func (g *ModelClient) toContent(entry llm.ChatEntry) *genai.Content {
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text(entry.Text),
		},
		Role: g.ToProviderRole(entry.Role),
	}
	return content
}

// The gen-ai-go API requires the last entry of a conversation to be submitted on its own
func extractLast(conversation llm.Conversation) ([]llm.ChatEntry, llm.ChatEntry) {
	return conversation.Entries[:len(conversation.Entries)-1], conversation.Entries[len(conversation.Entries)-1]
}

func (g *ModelClient) chatSession() *genai.ChatSession {
	if g.cachedChatSession == nil {
		generativeModel := g.generativeModel()
		g.cachedChatSession = generativeModel.StartChat()
	}
	return g.cachedChatSession
}

func (g *ModelClient) generativeModel() *genai.GenerativeModel {
	if g.cachedModel == nil {
		generativeModel := g.client.GenerativeModel(g.model)
		generativeModel.SafetySettings = harmBlockNone()
		generativeModel.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(g.config.String(keys.OptionSystemPrompt))},
		}
		g.cachedModel = generativeModel
	}
	return g.cachedModel
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
