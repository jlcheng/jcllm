package llm

import (
	"context"
	"errors"
	"iter"
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
		Role     string
		Messages iter.Seq2[Message, error]
	}

	Message struct {
		TokenCount int
		Text       string
	}
)

const (
	RoleSystem    = "RoleSystem"
	RoleUser      = "RoleUser"
	RoleAssistant = "RoleAssistant"
)

var ErrProviderNotFound = errors.New("provider not found")
var ErrBlankInput = errors.New("no input")
