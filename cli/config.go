package cli

import (
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
)

var ConfigMetadata = []configuration.Metadata{
	{keys.OptionProvider, keys.ProviderGeminiLegacy, "provider name"},
	{keys.OptionLogFile, "", "If specified, log to this diagnostic log file"},
	{keys.OptionModel, "gemini-1.5-flash-8b", "model name"},
	{keys.OptionGeminiApiKey, "", "Gemini API Key"},
	{keys.OptionCommand, "repl", "Supported commands are: list-models, list-providers, repl"},
	{keys.OptionSystemPrompt, "You are an AI assistant. Be concise.", "If specified, use this system prompt"},
}

var ConfigBools = []configuration.Metadata{
	{keys.OptionsVersion, "", "Show version information."},
}
