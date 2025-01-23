package cli

import (
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
)

var ConfigMetadata = []configuration.Metadata{
	{keys.OptionCommand, "repl", "Supported commands are: list-models, list-providers, repl"},
	{keys.OptionGeminiApiKey, "", "Gemini API Key"},
	{keys.OptionHttpTimeout, "30", "The http timeout, in seconds"},
	{keys.OptionLogFile, "", "If specified, log to this diagnostic log file"},
	{keys.OptionModel, "gemini-1.5-flash-8b", "model name"},
	{keys.OptionOpenAIApiKey, "", "OpenAI API key"},
	{keys.OptionOpenAIBaseURL, "https://api.openai.com/v1", "OpenAI base url, which could be replaced with an OpenAI-compatible base url, such as https://generativelanguage.googleapis.com/v1beta/openai"},
	{keys.OptionProvider, keys.ProviderOpenAI, "The LLM provider. Examples are: gemini and openai"},
	{keys.OptionSystemPrompt, "You are an AI assistant. Be concise.", "If specified, use this system prompt"},
}

var ConfigBools = []configuration.Metadata{
	{keys.OptionVersion, "", "Show version information."},
}
