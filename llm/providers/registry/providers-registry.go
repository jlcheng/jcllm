package registry

import (
	"context"
	"fmt"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/keys"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/llm/providers/googlegenai"
)

func NewProvider(ctx context.Context, configuration configuration.Configuration, name string) (llm.ProviderIfc, error) {
	switch name {
	case keys.ProviderGemini:
		return googlegenai.NewProvider(configuration), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
