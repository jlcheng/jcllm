package googlegenai

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/go-errors/errors"
	"google.golang.org/genai"
	"io"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/keys"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/log"
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

type Provider struct {
	config configuration.Configuration
	logger *log.Logger
}

// NewProvider creates a provider to models powered by https://pkg.go.dev/google.golang.org/genai.
func NewProvider(config configuration.Configuration) *Provider {
	return &Provider{
		config: config,
		logger: log.New(config.String(keys.OptionLogFile)),
	}
}

func (p *Provider) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	listModelURL := "https://generativelanguage.googleapis.com/v1beta/models"
	resp, err := http.Get(fmt.Sprintf("%s?key=%s", listModelURL, p.config.String(keys.OptionGeminiApiKey)))
	if err != nil {
		return nil, errors.WrapPrefix(err, "error getting model list", 0)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WrapPrefix(err, "error reading list-models response", 0)
	}
	var listModelsOutput ListModelsOutput
	if err := json.Unmarshal(body, &listModelsOutput); err != nil {
		return nil, errors.WrapPrefix(err, "json parse error", 0)
	}
	modelsList := slices.Collect(
		it.Map(slices.Values(listModelsOutput.Models), func(model ModelInfo) llm.ModelInfo {
			return llm.ModelInfo{
				DisplayName: model.DisplayName,
				Name:        model.Name,
				Description: model.Description,
				MaxTokens:   model.MaxTokens,
				Version:     model.Version,
			}
		}),
	)
	return modelsList, nil
}

func (p *Provider) ToProviderRole(genericRole string) (providerRole string) {
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

func (p *Provider) ToGenericRole(providerRole string) (genericRole string) {
	switch providerRole {
	case RoleModel:
		return llm.RoleAssistant
	case RoleUser:
		return llm.RoleUser
	}
	return genericRole
}

func (p *Provider) SolicitResponse(ctx context.Context, input llm.SolicitResponseInput) (llm.ResponseStream, error) {
	conversation := input.Conversation
	exchange := make(chan llm.Message)
	response := llm.ResponseStream{
		Role:           p.ToGenericRole(RoleModel),
		ResponseStream: exchange,
	}
	go func() {
		contents := slices.Collect(it.Map(slices.Values(conversation.Entries), func(v llm.ChatEntry) *genai.Content {
			return &genai.Content{
				Parts: []*genai.Part{{Text: v.Text}},
				Role:  p.ToProviderRole(v.Role),
			}
		}))
		sdkClient, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  p.config.String(keys.OptionGeminiApiKey),
			Backend: genai.BackendGoogleAI,
		})
		if err != nil {
			exchange <- llm.Message{Err: errors.WrapPrefix(err, "error getting gemini client", 0)}
		}
		tools := p.handleGroundingSupport(input, nil)

		groundingDataBuffer := new(strings.Builder)
		for chunk, err := range sdkClient.Models.GenerateContentStream(ctx, input.ModelName, contents, &genai.GenerateContentConfig{
			SystemInstruction: genai.Text(p.config.String(keys.OptionSystemPrompt))[0],
			Tools:             tools,
			SafetySettings:    harmBlockNone(),
		}) {
			if err != nil {
				exchange <- llm.Message{Err: errors.WrapPrefix(err, "failed to generate content", 0)}
				continue
			}
			if len(chunk.Candidates) == 0 {
				continue
			}
			resp := chunk.Candidates[0]
			if resp.FinishReason != "" && resp.FinishReason != genai.FinishReasonStop {
				exchange <- llm.Message{Err: errors.Errorf("model stopped: %s", resp.FinishReason)}
				continue
			}
			if resp.Content == nil {
				continue
			}
			if resp.GroundingMetadata != nil && resp.GroundingMetadata.GroundingChunks != nil {
				for _, chunk := range resp.GroundingMetadata.GroundingChunks {
					if chunk.Web != nil {
						fmt.Fprintf(groundingDataBuffer, "[%s](%s)\n", chunk.Web.Title, chunk.Web.URI)
					}
				}
			}
			buf := new(strings.Builder)
			for _, part := range resp.Content.Parts {
				if text, ok := mapToText(part); ok {
					buf.WriteString(text)
				} else {
					p.logger.Debugf("unknown part type: %+v", *part)
					exchange <- llm.Message{Err: fmt.Errorf("unknown part type: %s", part.Text)}
					break
				}
			}
			exchange <- llm.Message{
				TokenCount: getTokenCount(chunk),
				Text:       buf.String(),
				Err:        nil,
			}
			buf.Reset()
		}
		if groundingDataBuffer.Len() > 0 {
			exchange <- llm.Message{
				Text: fmt.Sprintf("\n\n%s", strings.TrimSpace(groundingDataBuffer.String())),
			}
		}
		close(exchange)
	}()
	return response, nil
}

func (p *Provider) isGroundingEnabled(conversation llm.Conversation) bool {
	if len(conversation.Entries) == 0 {
		return false
	}
	lastEntry := conversation.Entries[len(conversation.Entries)-1]
	text := strings.ToLower(lastEntry.Text)
	return strings.Contains(text, "use grounding")
}

func (p *Provider) handleGroundingSupport(input llm.SolicitResponseInput, tools []*genai.Tool) []*genai.Tool {
	conversation := input.Conversation
	if len(conversation.Entries) == 0 {
		return tools
	}
	lastEntry := conversation.Entries[len(conversation.Entries)-1]
	text := strings.ToLower(lastEntry.Text)
	if strings.Contains(text, "use grounding") || strings.Contains(text, "ground with search") {
		if tools == nil {
			tools = make([]*genai.Tool, 0)
		}
		var searchTool = &genai.Tool{}
		if strings.HasPrefix(input.ModelName, "gemini-2.0-flash") {
			searchTool.GoogleSearch = &genai.GoogleSearch{}
		} else {
			searchTool.GoogleSearchRetrieval = &genai.GoogleSearchRetrieval{}
		}
		return append(tools, searchTool)
	}
	return tools
}

func getTokenCount(chunk *genai.GenerateContentResponse) int {
	if chunk == nil || chunk.UsageMetadata == nil {
		return 0
	}
	return int(chunk.UsageMetadata.CandidatesTokenCount)
}

func mapToText(part *genai.Part) (string, bool) {
	if part.Text != "" {
		return part.Text, true
	} else if part.InlineData != nil {
		return fmt.Sprintf("(inline-data type: %s)\n", part.InlineData.MIMEType), true
	} else if part.FunctionResponse != nil {
		return fmt.Sprintf("(function-response name: %s id: %s)\n",
			part.FunctionResponse.Name, part.FunctionResponse.ID), true
	} else if part.FunctionCall != nil {
		return fmt.Sprintf("(function-call name: %s id: %s)\n",
			part.FunctionCall.Name, part.FunctionCall.ID), true
	} else if part.FileData != nil {
		return fmt.Sprintf("(file-data uri: %s)\n", part.FileData.FileURI), true
	} else if part.ExecutableCode != nil {
		return fmt.Sprintf("(executable-code lang: %s, code: %s)\n",
			part.ExecutableCode.Language, part.ExecutableCode.Code), true
	} else if part.CodeExecutionResult != nil {
		return fmt.Sprintf("(code-execution-result output: %s)\n", part.CodeExecutionResult.Output), true
	} else if part.VideoMetadata != nil {
		return fmt.Sprintf("(video-meta start: %s end: %s)\n",
			part.VideoMetadata.StartOffset, part.VideoMetadata.StartOffset), true
	}
	return "", false
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
			Threshold: genai.HarmBlockThresholdBlockNone,
		}
	}))
}

type ModelInfo struct {
	Name        string `json:"name"`
	BaseModelID string `json:"baseModelId"`
	Version     string `json:"version"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	MaxTokens   int    `json:"maxTokens"`
}

type ListModelsOutput struct {
	Models []ModelInfo `json:"models"`
}

var _ llm.ProviderIfc = (*Provider)(nil)
