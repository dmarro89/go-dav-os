package agent

type DeterministicPlanner struct{}

func (DeterministicPlanner) Plan(input string, _ *Context) PlanningResult {
	if contains(input, "help") || contains(input, "commands") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowHelp, ActionShowHelp, RiskSafe))
	}
	if contains(input, "ls") || contains(input, "files") || contains(input, "file") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentListFiles, ActionListFiles, RiskSafe))
	}
	return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowHelp, ActionShowHelp, RiskSafe))
}

type BridgeClient interface {
	Plan(input string, context *Context) PlanningResult
}

type LLMPlanner struct {
	Bridge BridgeClient
}

const errLLMBridgeNotConfigured = "agent: llm bridge not configured"

func (p LLMPlanner) Plan(input string, context *Context) PlanningResult {
	if p.Bridge == nil {
		return PlanningResult{OK: false, Reason: errLLMBridgeNotConfigured}
	}

	result := p.Bridge.Plan(input, context)
	if !result.OK {
		if result.Reason == "" {
			result.Reason = "agent: llm bridge failed"
		}
		return result
	}
	result.Plan.Planner = PlannerModeLLM
	return result
}

func singleActionPlan(mode PlannerMode, intent IntentKind, actionKind ActionKind, risk RiskLevel) Plan {
	var plan Plan
	plan.Planner = mode
	plan.Intent = intent
	plan.Confidence = 100
	plan.ActionCount = 1
	plan.Actions[0] = Action{Kind: actionKind, Risk: risk}
	return plan
}

func successfulPlan(plan Plan) PlanningResult {
	return PlanningResult{OK: true, Plan: plan}
}

func contains(text, token string) bool {
	if len(token) == 0 {
		return true
	}
	if len(token) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(token); i++ {
		match := true
		for j := 0; j < len(token); j++ {
			if lowerASCII(text[i+j]) != lowerASCII(token[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func lowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}
