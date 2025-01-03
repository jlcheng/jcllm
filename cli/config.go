package cli

import (
	"jcheng.org/jcllm/configuration"
)

var ConfigMetadata = []configuration.Metadata{
	{"provider", "openai", "provider name"},
	{"model", "gpt-4o-mini", "model name"},
	{"openai-api-key", "", "OpenAI API Key"},
	{"gemini-api-key", "", "Gemini API Key"},
	{"http-timeout", "5", "http timeout in seconds"},
	{"command", "", "Supported commands are: list-models, list-providers"},
}
