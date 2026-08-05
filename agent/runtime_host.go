//go:build !gccgo

package agent

func (r Runtime) Run(input string, context *Context) Response {
	mode := r.PlannerMode()
	if context != nil {
		context.BeginStringRequest(input, mode)
	}
	var planner Planner = DeterministicPlanner{}
	if mode == PlannerModeLLM {
		planner = r.llmPlanner
	}
	var planning PlanningResult
	if planner == nil || !planner.Available() {
		planning = PlanningResult{OK: false, Reason: MessageLLMBridgeNotConfigured}
	} else {
		planning = planner.Plan(input, context)
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
