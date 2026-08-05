//go:build !gccgo

package agent

func (r *Runtime) ConfigureLLMPlanner(planner Planner) {
	if r == nil {
		return
	}
	if planner == nil || !planner.Available() {
		r.llmConfigured = false
		r.llmPlan = nil
		if r.plannerMode == PlannerModeLLM {
			r.plannerMode = PlannerModeDeterministic
		}
		return
	}
	r.llmConfigured = true
	r.llmPlan = planner.Plan
}

func (r Runtime) Run(input string, context *Context) Response {
	mode := r.PlannerMode()
	if context != nil {
		context.BeginStringRequest(input, mode)
	}
	var planning PlanningResult
	if mode == PlannerModeLLM && r.llmPlan == nil {
		planning = PlanningResult{OK: false, Reason: MessageLLMBridgeNotConfigured}
	} else if mode == PlannerModeLLM {
		planning = r.llmPlan(input, context)
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
