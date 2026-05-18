package agent

type Runtime struct {
	Executor           AllowedActionExecutor
	ExecutorConfigured bool
}

func NewDeterministicAgent(executor AllowedActionExecutor) Runtime {
	var runtime Runtime
	runtime.Executor.ListFiles = executor.ListFiles
	runtime.Executor.ReadFile = executor.ReadFile
	runtime.Executor.WriteFile = executor.WriteFile
	runtime.Executor.DeleteFile = executor.DeleteFile
	runtime.Executor.StatFile = executor.StatFile
	runtime.Executor.ShowHelp = executor.ShowHelp
	runtime.Executor.ShowHistory = executor.ShowHistory
	runtime.Executor.SetMode = executor.SetMode
	runtime.ExecutorConfigured = true
	return runtime
}

func (r Runtime) RunAction(kind ActionKind, intent IntentKind, risk RiskLevel, target *[MaxNameLen]byte, targetLen int, context *Context) Response {
	plan := singleActionPlan(PlannerModeDeterministic, intent, kind, risk)
	if target != nil && targetLen > 0 {
		if targetLen > MaxNameLen {
			targetLen = MaxNameLen
		}
		plan.Actions[0].TargetLen = targetLen
		for i := 0; i < targetLen; i++ {
			plan.Actions[0].Target[i] = target[i]
		}
	}
	return r.runPlan(plan, context)
}

func (r *Runtime) RunActionMessage(kind ActionKind, intent IntentKind, risk RiskLevel, target *[MaxNameLen]byte, targetLen int, context *Context) MessageKind {
	if !kind.Valid() {
		return MessagePlanContainsUnsupportedAction
	}
	if !validRiskLevel(risk) {
		return MessageActionRiskInvalid
	}
	if targetLen < 0 || targetLen > MaxNameLen {
		return MessageActionTargetInvalid
	}
	if risk == RiskRisky {
		return MessageConfirmationRequired
	}
	if r == nil || !r.ExecutorConfigured {
		return MessageExecutorNotConfigured
	}

	var action Action
	action.Kind = kind
	action.Risk = risk
	if target != nil && targetLen > 0 {
		action.TargetLen = targetLen
		for i := 0; i < targetLen; i++ {
			action.Target[i] = target[i]
		}
	}

	result := r.Executor.Execute(action, context)
	if result.OK && context != nil {
		context.Remember(intent)
	}
	return result.Message
}

func (r Runtime) runPlan(plan Plan, context *Context) Response {
	var response Response
	response.AddTrace(TracePlanner, plannerTrace(plan.Planner))
	response.AddTrace(TraceIntent, intentTrace(plan.Intent))

	validation := validatePlan(plan)
	if !validation.OK {
		setResponseResult(&response, false, validation.Reason)
		setSafety(&response, SafetyRejected, MessageValidationFailed)
		response.AddTrace(TraceValidation, traceFromMessage(validation.Reason))
		return response
	}
	response.AddTrace(TraceValidation, TraceDetailOK)

	safety := evaluateSafety(plan, context)
	setSafety(&response, safety.Status, safety.Reason)
	response.AddTrace(TraceSafety, safetyTrace(safety.Status))
	if safety.Status != SafetyAllowed {
		setResponseResult(&response, false, safety.Reason)
		return response
	}

	if !r.ExecutorConfigured {
		setResponseResult(&response, false, MessageExecutorNotConfigured)
		response.AddTrace(TraceExecutor, TraceDetailMissing)
		return response
	}

	var results [MaxActions]ActionResult
	for i := 0; i < plan.ActionCount; i++ {
		results[i] = r.Executor.Execute(plan.Actions[i], context)
		if !results[i].OK {
			response.AddTrace(TraceExecutor, TraceDetailFailed)
			setResponseResult(&response, results[i].OK, results[i].Message)
			return response
		}
	}
	response.AddTrace(TraceExecutor, TraceDetailSuccess)

	result := formatResult(plan, results, plan.ActionCount, safety)
	setResponseResult(&response, result.OK, result.Message)
	response.AddTrace(TraceFormatter, TraceDetailStructured)
	if context != nil {
		context.Remember(plan.Intent)
	}
	return response
}

func setResponseResult(response *Response, ok bool, message MessageKind) {
	response.Result.OK = ok
	response.Result.Message = message
}

func setSafety(response *Response, status SafetyStatus, reason MessageKind) {
	response.Safety.Status = status
	response.Safety.Reason = reason
}

func validatePlan(plan Plan) ValidationResult {
	if plan.ActionCount <= 0 {
		return ValidationResult{OK: false, Reason: MessagePlanHasNoActions}
	}
	if plan.ActionCount > MaxActions {
		return ValidationResult{OK: false, Reason: MessagePlanHasTooManyActions}
	}
	for i := 0; i < plan.ActionCount; i++ {
		action := plan.Actions[i]
		if !action.Kind.Valid() {
			return ValidationResult{OK: false, Reason: MessagePlanContainsUnsupportedAction}
		}
		if !validRiskLevel(action.Risk) {
			return ValidationResult{OK: false, Reason: MessageActionRiskInvalid}
		}
		if action.TargetLen < 0 || action.TargetLen > MaxNameLen {
			return ValidationResult{OK: false, Reason: MessageActionTargetInvalid}
		}
		if action.DataLen < 0 || action.DataLen > MaxDataLen {
			return ValidationResult{OK: false, Reason: MessageActionDataInvalid}
		}
	}
	return ValidationResult{OK: true, Reason: MessageOK}
}

func validRiskLevel(risk RiskLevel) bool {
	switch risk {
	case RiskSafe, RiskRisky:
		return true
	default:
		return false
	}
}

func evaluateSafety(plan Plan, _ *Context) SafetyDecision {
	for i := 0; i < plan.ActionCount; i++ {
		if plan.Actions[i].Risk == RiskRisky {
			return SafetyDecision{Status: SafetyConfirmationRequired, Reason: MessageConfirmationRequired}
		}
	}
	return SafetyDecision{Status: SafetyAllowed, Reason: MessageOK}
}

func formatResult(_ Plan, results [MaxActions]ActionResult, resultCount int, _ SafetyDecision) ActionResult {
	if resultCount <= 0 {
		return ActionResult{OK: false, Message: MessageNoResult}
	}
	if resultCount == 1 {
		return results[0]
	}
	return ActionResult{OK: true, Message: MessageCompletedPlan}
}

func plannerTrace(mode PlannerMode) TraceDetail {
	if mode == PlannerModeLLM {
		return TraceDetailLLM
	}
	return TraceDetailDeterministic
}

func intentTrace(intent IntentKind) TraceDetail {
	switch intent {
	case IntentListFiles:
		return TraceDetailListFiles
	case IntentReadFile:
		return TraceDetailReadFile
	case IntentWriteFile:
		return TraceDetailWriteFile
	case IntentDeleteFile:
		return TraceDetailDeleteFile
	case IntentStatFile:
		return TraceDetailStatFile
	case IntentShowHelp:
		return TraceDetailShowHelp
	case IntentShowHistory:
		return TraceDetailShowHistory
	case IntentSetMode:
		return TraceDetailSetMode
	default:
		return TraceDetailNone
	}
}

func safetyTrace(status SafetyStatus) TraceDetail {
	switch status {
	case SafetyAllowed:
		return TraceDetailAllowed
	case SafetyConfirmationRequired:
		return TraceDetailConfirmationRequired
	default:
		return TraceDetailRejected
	}
}

func traceFromMessage(message MessageKind) TraceDetail {
	if message == MessageOK {
		return TraceDetailOK
	}
	return TraceDetailFailed
}
