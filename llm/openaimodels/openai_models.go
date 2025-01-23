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
