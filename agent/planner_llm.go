//go:build !gccgo

package agent

type BridgeClient interface {
	Plan(input string, context *Context) PlanningResult
}

type LLMPlanner struct {
	Bridge BridgeClient
}

const errLLMBridgeNotConfigured = "agent: llm bridge not configured"

func (p LLMPlanner) Plan(input string, context *Context) PlanningResult {
	if p.Bridge == nil {
		return PlanningResult{OK: false, Reason: MessageLLMBridgeNotConfigured}
	}

	result := p.Bridge.Plan(input, context)
	if !result.OK {
		if result.Reason == MessageNone {
			result.Reason = MessageLLMBridgeFailed
		}
		return result
	}
	result.Plan.Planner = PlannerModeLLM
	return result
}
