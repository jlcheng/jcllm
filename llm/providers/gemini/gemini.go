package gemini

import (
	"context"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
)

const (
	// RoleUser is 'user'
	RoleUser = "user"
	// RoleModel is  'model', as Gemini only recognize 'user' or 'model'.
	RoleModel = "model"
)

// Gemini implements a llm.ModelIfc using the Vertex AI Go SDK
type Gemini struct {
	config configuration.Configuration
}

func New(config configuration.Configuration) *Gemini {
	return &Gemini{config: config}
}

func (g *Gemini) ToProviderRole(genericRole string) (providerRole string) {
	switch genericRole {
	case llm.RoleAssistant:
		return RoleModel
	case llm.RoleUser:
		return RoleUser
	case llm.RolesSystem:
		return RoleUser
	}
	return genericRole
}

func (g *Gemini) GetModel(ctx context.Context, modelName string) (llm.ModelIfc, error) {
	return nil, llm.ErrNoSupport
}

func (g *Gemini) ToGenericRole(providerRole string) (genericRole string) {
	switch providerRole {
	case RoleModel:
		return llm.RoleAssistant
	case RoleUser:
		return llm.RoleUser
	}
	return genericRole
}

func (g *Gemini) ListModels(context.Context) ([]llm.ModelInfo, error) {
	return nil, llm.ErrNoSupport
}

func (g *Gemini) SolicitResponse(context.Context, llm.Conversation) (llm.ResponseStream, error) {
	return llm.ResponseStream{}, llm.ErrNoSupport
}

var _ llm.ModelIfc = (*Gemini)(nil)
var _ llm.ProviderIfc = (*Gemini)(nil)
var _ llm.RoleMapper = (*Gemini)(nil)
