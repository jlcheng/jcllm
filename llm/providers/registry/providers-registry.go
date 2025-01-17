package registry

import (
	"context"
	"fmt"
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
	"github.com/jlcheng/jcllm/llm"
	"github.com/jlcheng/jcllm/llm/providers/googlegenai"
)

func NewProvider(ctx context.Context, configuration configuration.Configuration, name string) (llm.ProviderIfc, error) {
	switch name {
	case keys.ProviderGemini:
		return googlegenai.NewProvider(configuration), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
