//go:build !gccgo

package agent

var configuredLLMPlanner Planner

func (r *Runtime) ConfigureLLMPlanner(planner Planner) {
	if r == nil {
		return
	}
	if planner == nil || !planner.Available() {
		configuredLLMPlanner = nil
		r.llmConfigured = false
		if r.plannerMode == PlannerModeLLM {
			r.plannerMode = PlannerModeDeterministic
		}
		return
	}
	configuredLLMPlanner = planner
	r.llmConfigured = true
}

func (r Runtime) Run(input string, context *Context) Response {
	mode := r.PlannerMode()
	if context != nil {
		context.BeginStringRequest(input, mode)
	}
	var planning PlanningResult
	if mode == PlannerModeLLM && configuredLLMPlanner == nil {
		planning = PlanningResult{OK: false, Reason: MessageLLMBridgeNotConfigured}
	} else if mode == PlannerModeLLM {
		planning = configuredLLMPlanner.Plan(input, context)
	} else {
		planning = deterministicPlan(input, context)
	}
	if !planning.OK {
		var response Response
		setResponseResult(&response, false, planning.Reason)
		setSafety(&response, SafetyRejected, MessagePlannerFailed)
		response.AddTrace(TracePlanner, traceFromMessage(planning.Reason))
		if context != nil {
			context.CurrentTask = ActionNone
			context.LastIntent = IntentUnknown
			context.LastAction = ActionNone
			context.LastResultSummary = planning.Reason
		}
		return response
	}
	return r.runPlan(planning.Plan, context)
}
