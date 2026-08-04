//go:build !gccgo

package agent

func (r Runtime) Run(input string, context *Context) Response {
	if context != nil {
		context.BeginStringRequest(input, PlannerModeDeterministic)
	}
	planning := deterministicPlan(input, context)
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
