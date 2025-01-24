// Package openaimodels provides json object definitions for objects described in OpenAI documentations.
package openaimodels

type ListModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// CreateChatCompletionRequest represents the request body for the "Create chat completion" API.
type CreateChatCompletionRequest struct {
	Model           string          `json:"model"`
	Messages        []Message       `json:"messages"`
	ReasoningEffort *string         `json:"reasoning_effort,omitempty"`
	ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
	Stream          *bool           `json:"stream,omitempty"`
	StreamOptions   *StreamOptions  `json:"stream_options,omitempty"`
}

// Message represents a message in the messages array.
type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// ResponseFormat specifies the format that the model must output.
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema describes the json_schema object for structured outputs.
type JSONSchema struct {
	Description *string     `json:"description,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Schema      interface{} `json:"schema,omitempty"` // Represents the schema object, type is not strictly defined in the documentation.
	Strict      *bool       `json:"strict,omitempty"`
}

// StreamOptions provides options for streaming responses.
type StreamOptions struct {
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

type ChatCompletionResponse struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	SystemFingerprint string       `json:"system_fingerprint"`
	Choices           []ChatChoice `json:"choices"`
	ServiceTier       string       `json:"service_tier"`
	Usage             ChatUsage    `json:"usage"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	Logprobs     interface{} `json:"logprobs"` // Can be null
	FinishReason string      `json:"finish_reason"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatUsage struct {
	PromptTokens            int                    `json:"prompt_tokens"`
	CompletionTokens        int                    `json:"completion_tokens"`
	TotalTokens             int                    `json:"total_tokens"`
	CompletionTokensDetails CompletionTokenDetails `json:"completion_tokens_details"`
}

type CompletionTokenDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type ChatCompletionChunkResponse struct {
	ID                string            `json:"id"`
	Object            string            `json:"object"`
	Created           int64             `json:"created"`
	Model             string            `json:"model"`
	SystemFingerprint string            `json:"system_fingerprint"`
	Choices           []ChatChunkChoice `json:"choices"`
	Usage             *ChatUsage        `json:"usage"`
}

type ChatChunkChoice struct {
	Index        int         `json:"index"`
	Delta        ChatDelta   `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`      // Can be null
	FinishReason *string     `json:"finish_reason"` // Can be null
}

type ChatDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type APIErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Message string `json:"message"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}
