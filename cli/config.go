package cli

import (
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/keys"
)

var ConfigMetadata = []configuration.Metadata{
	{keys.OptionProvider, keys.ProviderGemini, "provider name"},
	{keys.OptionModel, "gemini-1.5-flash-8b", "model name"},
	{keys.OptionGeminiApiKey, "", "Gemini API Key"},
	{keys.OptionCommand, "repl", "Supported commands are: list-models, list-providers, repl"},
	{keys.OptionSystemPrompt, "You are an AI assistant. Be concise.", "System prompt"},
}
