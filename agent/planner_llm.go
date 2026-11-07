//go:build !gccgo

package agent

type BridgeClient interface {
	Available() bool
	Plan(request BridgeRequest) BridgeResult
}

type LLMPlanner struct {
	Bridge BridgeClient
}

var _ Planner = LLMPlanner{}

const errLLMBridgeNotConfigured = "agent: llm bridge not configured"

func (p LLMPlanner) Available() bool {
	return p.Bridge != nil && p.Bridge.Available()
}

func (p LLMPlanner) Plan(input string, context *Context) PlanningResult {
	if p.Bridge == nil {
		return PlanningResult{OK: false, Reason: MessageLLMBridgeNotConfigured}
	}

	var requestInput [MaxContextInput]byte
	inputLen := len(input)
	if inputLen > MaxContextInput {
		inputLen = MaxContextInput
	}
	for i := 0; i < inputLen; i++ {
		requestInput[i] = input[i]
	}
	request := NewBridgeRequest(&requestInput, inputLen, context)
	return bridgePlanningResult(request, p.Bridge.Plan(request))
}
