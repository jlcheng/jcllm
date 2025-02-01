package googlegenai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/go-errors/errors"
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
	"github.com/jlcheng/jcllm/extract"
	"github.com/jlcheng/jcllm/llm"
	"github.com/jlcheng/jcllm/log"
	"google.golang.org/genai"
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
				Name:        strings.TrimPrefix(model.Name, "models/"),
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
	case llm.RoleSystem:
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
	sdkClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  p.config.String(keys.OptionGeminiApiKey),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return llm.ResponseStream{}, errors.WrapPrefix(err, "gemini client creation failed", 0)
	}
	response := llm.ResponseStream{
		Role: p.ToGenericRole(RoleModel),
	}
	tools := p.handleGroundingSupport(input, nil)
	contents := slices.Collect(it.Map(slices.Values(conversation.Entries), func(v llm.ChatEntry) *genai.Content {
		return &genai.Content{
			Parts: []*genai.Part{{Text: v.Text}},
			Role:  p.ToProviderRole(v.Role),
		}
	}))
	sdkResponse := sdkClient.Models.GenerateContentStream(ctx, input.ModelName, contents, &genai.GenerateContentConfig{
		SystemInstruction: genai.Text(p.config.String(keys.OptionSystemPrompt))[0],
		Tools:             tools,
		SafetySettings:    harmBlockNone(),
	})
	response.Messages = it.Map2(sdkResponse, func(chunk *genai.GenerateContentResponse, err error) (llm.Message, error) {
		if err != nil {
			return llm.Message{}, errors.WrapPrefix(err, "generate content failed", 0)
		}
		if chunk == nil || chunk.Candidates == nil || len(chunk.Candidates) == 0 {
			return llm.Message{}, nil
		}
		resp := chunk.Candidates[0]
		if resp.FinishReason != "" && resp.FinishReason != genai.FinishReasonStop {
			return llm.Message{}, errors.Errorf("model stopped: %s", resp.FinishReason)
		}

		buf := new(strings.Builder)
		for _, part := range resp.Content.Parts {
			if text, ok := mapToText(part); ok {
				buf.WriteString(text)
			} else {
				p.logger.Debugf("unknown part type: %+v", *part)
				return llm.Message{}, errors.Errorf("unknown part type: %+v", *part)
			}
		}

		if resp.GroundingMetadata != nil && resp.GroundingMetadata.GroundingChunks != nil && len(resp.GroundingMetadata.GroundingChunks) != 0 {
			mustFprintf(buf, "Citations:\n")
			for _, chunk := range resp.GroundingMetadata.GroundingChunks {
				if chunk.Web != nil {
					mustFprintf(buf, "  [%s](%s)\n", chunk.Web.Title, chunk.Web.URI)
				}
			}
		}

		return llm.Message{
			TokenCount: getTokenCount(chunk),
			Text:       buf.String(),
		}, nil
	})
	return response, nil
}

func (p *Provider) handleGroundingSupport(input llm.SolicitResponseInput, tools []*genai.Tool) []*genai.Tool {
	conversation := input.Conversation
	if len(conversation.Entries) == 0 || input.Args[keys.ArgNameSuppress] == keys.True {
		return tools
	}
	groundWithSearch := false
	lastEntry := &conversation.Entries[len(conversation.Entries)-1]
	newText, mentions := extract.MentionsFromEnd(lastEntry.Text)
	for _, mention := range mentions {
		if mention == "ground" {
			groundWithSearch = true
			break
		}
	}
	lastEntry.Text = newText
	if groundWithSearch {
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
	if chunk == nil || chunk.UsageMetadata == nil || chunk.UsageMetadata.CandidatesTokenCount == nil {
		return 0
	}
	return int(*chunk.UsageMetadata.CandidatesTokenCount)
}

func mapToText(part *genai.Part) (string, bool) {
	if part.InlineData != nil {
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
	return part.Text, true
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

func mustFprintf(w io.Writer, format string, args ...interface{}) {
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		panic(err)
	}
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
