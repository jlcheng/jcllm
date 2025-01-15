package llm

import (
	"context"
	"errors"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . ModelIfc
//counterfeiter:generate . RoleMapper
type (

	// ProviderIfc provides an API for interacting with model providers.
	ProviderIfc interface {
		RoleMapper
		ListModels(ctx context.Context) ([]ModelInfo, error)
		SolicitResponse(ctx context.Context, input SolicitResponseInput) (ResponseStream, error)
	}

	// RoleMapper maps the generic role to a provider-specific role and vice versa.
	RoleMapper interface {
		ToProviderRole(genericRole string) (providerRole string)
		ToGenericRole(providerRole string) (genericRole string)
	}
)

type (
	Conversation struct {
		Entries []ChatEntry `json:"entries"`
	}

	ChatEntry struct {
		Role string `json:"role"`
		Text string `json:"text"`
	}

	SolicitResponseInput struct {
		Conversation Conversation
		ModelName    string
		Args         map[string]string
	}

	ModelInfo struct {
		DisplayName string
		Name        string
		Description string
		MaxTokens   int
		Version     string
	}

	ResponseStream struct {
		Role           string
		ResponseStream <-chan Message
	}

	Message struct {
		TokenCount int
		Text       string
		Err        error
	}
)

const (
	RolesSystem   = "RoleSystem"
	RoleUser      = "RoleUser"
	RoleAssistant = "RoleAssistant"
)

var ErrNoSupport = errors.New("this model does not support the requested feature")
var ErrProviderNotFound = errors.New("provider not found")
var ErrModelNotFound = errors.New("model not found")
var ErrAPIKeyInvalid = errors.New("api key invalid")
