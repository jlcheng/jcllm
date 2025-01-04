package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/BooleanCat/go-functional/v2/it/itx"
	"github.com/go-errors/errors"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"net/http"
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
	config configuration.Configuration
	client *genai.Client
	model  string
}

type GoogleAPIErrorRoot struct {
	Error GoogleAPIError `json:"error"`
}

type GoogleAPIError struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Status  string           `json:"status"`
	Details []map[string]any `json:"details"`
}

func (errRoot GoogleAPIErrorRoot) IsEmpty() bool {
	return errRoot.Error.Status == ""
}

func (g *ModelClient) SolicitResponse(ctx context.Context, conversation llm.Conversation) (llm.ResponseStream, error) {
	client := g.client.GenerativeModel(g.model)
	client.SafetySettings = blockCategoriesNone
	chatSession := client.StartChat()
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
				exchange <- llm.Message{Err: g.mapStreamReadError(err)}
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
	return &ModelClient{g.config, g.client, model}, nil
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

func isModelNotFoundError(gerr *googleapi.Error) bool {
	return gerr.Code == http.StatusNotFound
}

func isNotAuthorizedError(gerr *googleapi.Error) bool {
	errList := unmarshalErrorBody(gerr)
	_, apiKeyInvalid := itx.FromSlice(errList).Find(func(elem GoogleAPIErrorRoot) bool {
		_, ok := itx.FromSlice(elem.Error.Details).Find(func(m map[string]any) bool {
			return m["reason"] == "API_KEY_INVALID"
		})
		return ok
	})
	return apiKeyInvalid
}

func unmarshalErrorBody(gerr *googleapi.Error) []GoogleAPIErrorRoot {
	var errList []GoogleAPIErrorRoot
	_ = json.Unmarshal([]byte(gerr.Body), &errList)
	return errList
}

func (g *ModelClient) mapStreamReadError(err error) error {
	var gerr *googleapi.Error
	// If not a googleapi.Error, simply return it
	if !errors.As(err, &gerr) {
		return err
	}

	// Get details out of Google API errors
	if isModelNotFoundError(gerr) {
		errorMessage := fmt.Sprintf("model [%s] not found", g.model)
		return errors.WrapPrefix(llm.ErrModelNotFound, errorMessage, 0)
	} else if isNotAuthorizedError(gerr) {
		return llm.ErrAPIKeyInvalid
	}
	errMsg := itx.FromSlice([]string{gerr.Message, gerr.Body, gerr.Error()}).Collect()[0]
	return errors.WrapPrefix(err, errMsg, 0)
}

var blockCategoriesNone = harmBlockNone()

var _ llm.ProviderIfc = (*Gemini)(nil)
var _ llm.ModelIfc = (*ModelClient)(nil)
var _ llm.RoleMapper = (*ModelClient)(nil)
