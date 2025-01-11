// Code generated by counterfeiter. DO NOT EDIT.
package llmfakes

import (
	"context"
	"sync"

	"jcheng.org/jcllm/llm"
)

type FakeModelIfc struct {
	ModelNameStub        func() string
	modelNameMutex       sync.RWMutex
	modelNameArgsForCall []struct {
	}
	modelNameReturns struct {
		result1 string
	}
	modelNameReturnsOnCall map[int]struct {
		result1 string
	}
	SolicitResponseStub        func(context.Context, llm.Conversation) (llm.ResponseStream, error)
	solicitResponseMutex       sync.RWMutex
	solicitResponseArgsForCall []struct {
		arg1 context.Context
		arg2 llm.Conversation
	}
	solicitResponseReturns struct {
		result1 llm.ResponseStream
		result2 error
	}
	solicitResponseReturnsOnCall map[int]struct {
		result1 llm.ResponseStream
		result2 error
	}
	ToGenericRoleStub        func(string) string
	toGenericRoleMutex       sync.RWMutex
	toGenericRoleArgsForCall []struct {
		arg1 string
	}
	toGenericRoleReturns struct {
		result1 string
	}
	toGenericRoleReturnsOnCall map[int]struct {
		result1 string
	}
	ToProviderRoleStub        func(string) string
	toProviderRoleMutex       sync.RWMutex
	toProviderRoleArgsForCall []struct {
		arg1 string
	}
	toProviderRoleReturns struct {
		result1 string
	}
	toProviderRoleReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeModelIfc) ModelName() string {
	fake.modelNameMutex.Lock()
	ret, specificReturn := fake.modelNameReturnsOnCall[len(fake.modelNameArgsForCall)]
	fake.modelNameArgsForCall = append(fake.modelNameArgsForCall, struct {
	}{})
	stub := fake.ModelNameStub
	fakeReturns := fake.modelNameReturns
	fake.recordInvocation("ModelName", []interface{}{})
	fake.modelNameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeModelIfc) ModelNameCallCount() int {
	fake.modelNameMutex.RLock()
	defer fake.modelNameMutex.RUnlock()
	return len(fake.modelNameArgsForCall)
}

func (fake *FakeModelIfc) ModelNameCalls(stub func() string) {
	fake.modelNameMutex.Lock()
	defer fake.modelNameMutex.Unlock()
	fake.ModelNameStub = stub
}

func (fake *FakeModelIfc) ModelNameReturns(result1 string) {
	fake.modelNameMutex.Lock()
	defer fake.modelNameMutex.Unlock()
	fake.ModelNameStub = nil
	fake.modelNameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeModelIfc) ModelNameReturnsOnCall(i int, result1 string) {
	fake.modelNameMutex.Lock()
	defer fake.modelNameMutex.Unlock()
	fake.ModelNameStub = nil
	if fake.modelNameReturnsOnCall == nil {
		fake.modelNameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.modelNameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeModelIfc) SolicitResponse(arg1 context.Context, arg2 llm.Conversation) (llm.ResponseStream, error) {
	fake.solicitResponseMutex.Lock()
	ret, specificReturn := fake.solicitResponseReturnsOnCall[len(fake.solicitResponseArgsForCall)]
	fake.solicitResponseArgsForCall = append(fake.solicitResponseArgsForCall, struct {
		arg1 context.Context
		arg2 llm.Conversation
	}{arg1, arg2})
	stub := fake.SolicitResponseStub
	fakeReturns := fake.solicitResponseReturns
	fake.recordInvocation("SolicitResponse", []interface{}{arg1, arg2})
	fake.solicitResponseMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeModelIfc) SolicitResponseCallCount() int {
	fake.solicitResponseMutex.RLock()
	defer fake.solicitResponseMutex.RUnlock()
	return len(fake.solicitResponseArgsForCall)
}

func (fake *FakeModelIfc) SolicitResponseCalls(stub func(context.Context, llm.Conversation) (llm.ResponseStream, error)) {
	fake.solicitResponseMutex.Lock()
	defer fake.solicitResponseMutex.Unlock()
	fake.SolicitResponseStub = stub
}

func (fake *FakeModelIfc) SolicitResponseArgsForCall(i int) (context.Context, llm.Conversation) {
	fake.solicitResponseMutex.RLock()
	defer fake.solicitResponseMutex.RUnlock()
	argsForCall := fake.solicitResponseArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeModelIfc) SolicitResponseReturns(result1 llm.ResponseStream, result2 error) {
	fake.solicitResponseMutex.Lock()
	defer fake.solicitResponseMutex.Unlock()
	fake.SolicitResponseStub = nil
	fake.solicitResponseReturns = struct {
		result1 llm.ResponseStream
		result2 error
	}{result1, result2}
}

func (fake *FakeModelIfc) SolicitResponseReturnsOnCall(i int, result1 llm.ResponseStream, result2 error) {
	fake.solicitResponseMutex.Lock()
	defer fake.solicitResponseMutex.Unlock()
	fake.SolicitResponseStub = nil
	if fake.solicitResponseReturnsOnCall == nil {
		fake.solicitResponseReturnsOnCall = make(map[int]struct {
			result1 llm.ResponseStream
			result2 error
		})
	}
	fake.solicitResponseReturnsOnCall[i] = struct {
		result1 llm.ResponseStream
		result2 error
	}{result1, result2}
}

func (fake *FakeModelIfc) ToGenericRole(arg1 string) string {
	fake.toGenericRoleMutex.Lock()
	ret, specificReturn := fake.toGenericRoleReturnsOnCall[len(fake.toGenericRoleArgsForCall)]
	fake.toGenericRoleArgsForCall = append(fake.toGenericRoleArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ToGenericRoleStub
	fakeReturns := fake.toGenericRoleReturns
	fake.recordInvocation("ToGenericRole", []interface{}{arg1})
	fake.toGenericRoleMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeModelIfc) ToGenericRoleCallCount() int {
	fake.toGenericRoleMutex.RLock()
	defer fake.toGenericRoleMutex.RUnlock()
	return len(fake.toGenericRoleArgsForCall)
}

func (fake *FakeModelIfc) ToGenericRoleCalls(stub func(string) string) {
	fake.toGenericRoleMutex.Lock()
	defer fake.toGenericRoleMutex.Unlock()
	fake.ToGenericRoleStub = stub
}

func (fake *FakeModelIfc) ToGenericRoleArgsForCall(i int) string {
	fake.toGenericRoleMutex.RLock()
	defer fake.toGenericRoleMutex.RUnlock()
	argsForCall := fake.toGenericRoleArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeModelIfc) ToGenericRoleReturns(result1 string) {
	fake.toGenericRoleMutex.Lock()
	defer fake.toGenericRoleMutex.Unlock()
	fake.ToGenericRoleStub = nil
	fake.toGenericRoleReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeModelIfc) ToGenericRoleReturnsOnCall(i int, result1 string) {
	fake.toGenericRoleMutex.Lock()
	defer fake.toGenericRoleMutex.Unlock()
	fake.ToGenericRoleStub = nil
	if fake.toGenericRoleReturnsOnCall == nil {
		fake.toGenericRoleReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.toGenericRoleReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeModelIfc) ToProviderRole(arg1 string) string {
	fake.toProviderRoleMutex.Lock()
	ret, specificReturn := fake.toProviderRoleReturnsOnCall[len(fake.toProviderRoleArgsForCall)]
	fake.toProviderRoleArgsForCall = append(fake.toProviderRoleArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ToProviderRoleStub
	fakeReturns := fake.toProviderRoleReturns
	fake.recordInvocation("ToProviderRole", []interface{}{arg1})
	fake.toProviderRoleMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeModelIfc) ToProviderRoleCallCount() int {
	fake.toProviderRoleMutex.RLock()
	defer fake.toProviderRoleMutex.RUnlock()
	return len(fake.toProviderRoleArgsForCall)
}

func (fake *FakeModelIfc) ToProviderRoleCalls(stub func(string) string) {
	fake.toProviderRoleMutex.Lock()
	defer fake.toProviderRoleMutex.Unlock()
	fake.ToProviderRoleStub = stub
}

func (fake *FakeModelIfc) ToProviderRoleArgsForCall(i int) string {
	fake.toProviderRoleMutex.RLock()
	defer fake.toProviderRoleMutex.RUnlock()
	argsForCall := fake.toProviderRoleArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeModelIfc) ToProviderRoleReturns(result1 string) {
	fake.toProviderRoleMutex.Lock()
	defer fake.toProviderRoleMutex.Unlock()
	fake.ToProviderRoleStub = nil
	fake.toProviderRoleReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeModelIfc) ToProviderRoleReturnsOnCall(i int, result1 string) {
	fake.toProviderRoleMutex.Lock()
	defer fake.toProviderRoleMutex.Unlock()
	fake.ToProviderRoleStub = nil
	if fake.toProviderRoleReturnsOnCall == nil {
		fake.toProviderRoleReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.toProviderRoleReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeModelIfc) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.modelNameMutex.RLock()
	defer fake.modelNameMutex.RUnlock()
	fake.solicitResponseMutex.RLock()
	defer fake.solicitResponseMutex.RUnlock()
	fake.toGenericRoleMutex.RLock()
	defer fake.toGenericRoleMutex.RUnlock()
	fake.toProviderRoleMutex.RLock()
	defer fake.toProviderRoleMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeModelIfc) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ llm.ModelIfc = new(FakeModelIfc)
