package registry

import (
	"context"
	"fmt"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/llm/providers/gemini"
)

func NewProvider(ctx context.Context, configuration configuration.Configuration, name string) (llm.ProviderIfc, error) {
	switch name {
	case "google":
		return gemini.NewProvider(ctx, configuration)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
