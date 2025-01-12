package registry

import (
	"context"
	"fmt"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/keys"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/llm/providers/gemini"
	"jcheng.org/jcllm/llm/providers/googlegenai"
)

func NewProvider(ctx context.Context, configuration configuration.Configuration, name string) (llm.ProviderIfc, error) {
	switch name {
	case keys.ProviderGemini:
		return gemini.NewProvider(ctx, configuration)
	case keys.ProviderGoogleAI:
		return googlegenai.NewProvider(configuration), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
