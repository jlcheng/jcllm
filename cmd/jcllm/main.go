package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"jcheng.org/jcllm/configuration"
	configProviderV1 "jcheng.org/jcllm/configuration/providers/clienvfile"
	"os"
)

func main() {
	configMetadata := []configuration.Metadata{
		{"provider", "openai", "provider name"},
		{"model", "gpt-4o-mini", "model name"},
		{"openai-api-key", "", "OpenAI API Key"},
		{"gemini-api-key", "", "Gemini API Key"},
		{"http-timeout", "5", "http timeout in seconds"},
	}
	cfg, err := configProviderV1.New(configMetadata)
	if err != nil {
		if errors.Is(err, configuration.ErrHelp) {
			os.Exit(0)
		}
		fmt.Println(err)
		os.Exit(1)
	}
	_ = cfg
	timeout := cfg.Int("http-timeout")
	fmt.Println(timeout)
}
