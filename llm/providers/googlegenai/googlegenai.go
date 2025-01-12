package googlegenai

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/go-errors/errors"
	"google.golang.org/genai"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/keys"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/log"
	"slices"
)

const (
	// RoleUser is 'user'
	RoleUser = "user"
	// RoleModel is  'model', as Gemini only recognize 'user' or 'model'.
	RoleModel = "model"
)

type Provider struct {
	config      configuration.Configuration
	logger      *log.Logger
	cachedModel *Model
}

// NewProvider creates a provider to models powered by https://pkg.go.dev/google.golang.org/genai.
func NewProvider(config configuration.Configuration) *Provider {
	return &Provider{
		config: config,
		logger: log.New(config.String(keys.OptionLogFile)),
	}
}

func (p *Provider) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return nil, llm.ErrNoSupport
}

func (p *Provider) GetModel(ctx context.Context, modelName string) (llm.ModelIfc, error) {
	if p.cachedModel == nil {
		sdkClient, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  p.config.String(keys.OptionGeminiApiKey),
			Backend: genai.BackendGoogleAI,
		})
		if err != nil {
			return nil, errors.WrapPrefix(err, "failed to initialize client from google.golang.org/genai", 0)
		}
		p.cachedModel = NewModel(modelName, p.config, sdkClient)
	}
	return p.cachedModel, nil
}

type Model struct {
	modelName string
	config    configuration.Configuration
	sdkClient *genai.Client
}

// NewModel returns a Model which can address requests for any model. It can be safely shared between models.
func NewModel(modelName string, config configuration.Configuration, sdkClient *genai.Client) *Model {
	return &Model{
		modelName: modelName,
		config:    config,
		sdkClient: sdkClient,
	}
}

func (m *Model) ToProviderRole(genericRole string) (providerRole string) {
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

func (m *Model) ToGenericRole(providerRole string) (genericRole string) {
	switch providerRole {
	case RoleModel:
		return llm.RoleAssistant
	case RoleUser:
		return llm.RoleUser
	}
	return genericRole
}

func (m *Model) ModelName() string {
	return m.modelName
}

func (m *Model) SolicitResponse(ctx context.Context, conversation llm.Conversation) (llm.ResponseStream, error) {
	history, lastEntry := extractLast(conversation)
	_ = history
	exchange := make(chan llm.Message)
	response := llm.ResponseStream{
		Role:           m.ToGenericRole(RoleModel),
		ResponseStream: exchange,
	}
	go func() {
		for sdkResponse, err := range m.sdkClient.Models.GenerateContentStream(ctx, m.modelName, genai.Text(lastEntry.Text), &genai.GenerateContentConfig{
			SystemInstruction: genai.Text(m.config.String(keys.OptionSystemPrompt))[0],
			Tools: []*genai.Tool{{
				GoogleSearchRetrieval: &genai.GoogleSearchRetrieval{},
			}},
			SafetySettings: harmBlockNone(),
		}) {
			buf := new(bytes.Buffer)
			flushBuffer := func() {
				if buf.Len() != 0 {
					exchange <- llm.Message{
						TokenCount: int(sdkResponse.UsageMetadata.CandidatesTokenCount),
						Text:       buf.String(),
						Err:        nil,
					}
				}
				buf.Reset()
			}
			if err != nil {
				flushBuffer()
				exchange <- llm.Message{Err: errors.WrapPrefix(err, "failed to generate content", 0)}
				continue
			}
			if len(sdkResponse.Candidates) == 0 {
				continue
			}
			finishReason := sdkResponse.Candidates[0].FinishReason
			if finishReason != "" && finishReason != genai.FinishReasonStop {
				flushBuffer()
				exchange <- llm.Message{Err: errors.Errorf("model stopped: %s", finishReason)}
				continue
			}
			if sdkResponse.Candidates[0].Content == nil {
				continue
			}
			for _, part := range sdkResponse.Candidates[0].Content.Parts {
				if part.Text != "" {
					buf.WriteString(part.Text)
				} else if part.InlineData != nil {
					buf.WriteString(fmt.Sprintf("(inline-data type: %s)\n", part.InlineData.MIMEType))
				} else if part.FunctionResponse != nil {
					buf.WriteString(fmt.Sprintf("(function-response name: %s id: %s)\n",
						part.FunctionResponse.Name, part.FunctionResponse.ID))
				} else if part.FunctionCall != nil {
					buf.WriteString(fmt.Sprintf("(function-call name: %s id: %s)\n",
						part.FunctionCall.Name, part.FunctionCall.ID))
				} else if part.FileData != nil {
					buf.WriteString(fmt.Sprintf("(file-data uri: %s)\n", part.FileData.FileURI))
				} else if part.ExecutableCode != nil {
					buf.WriteString(fmt.Sprintf("(executable-code lang: %s, code: %s)\n",
						part.ExecutableCode.Language, part.ExecutableCode.Code))
				} else if part.CodeExecutionResult != nil {
					buf.WriteString(fmt.Sprintf("(code-execution-result output: %s)\n", part.CodeExecutionResult.Output))
				} else if part.VideoMetadata != nil {
					buf.WriteString(fmt.Sprintf("(video-meta start: %s end: %s)\n",
						part.VideoMetadata.StartOffset, part.VideoMetadata.StartOffset))
				} else {
					exchange <- llm.Message{Err: fmt.Errorf("unknown part type: %s", part.Text)}
					break
				}
			}
			flushBuffer()
		}
		close(exchange)
	}()
	return response, nil
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

// The gen-ai-go API requires the last entry of a conversation to be submitted on its own
func extractLast(conversation llm.Conversation) ([]llm.ChatEntry, llm.ChatEntry) {
	return conversation.Entries[:len(conversation.Entries)-1], conversation.Entries[len(conversation.Entries)-1]
}

var _ llm.ProviderIfc = (*Provider)(nil)
var _ llm.ModelIfc = (*Model)(nil)
