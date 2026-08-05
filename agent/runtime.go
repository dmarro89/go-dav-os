package agent

type Runtime struct {
	Executor           AllowedActionExecutor
	ExecutorConfigured bool
	plannerMode        PlannerMode
	llmPlanner         Planner
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
	runtime.Executor.ShowVersion = executor.ShowVersion
	runtime.Executor.ShowTicks = executor.ShowTicks
	runtime.Executor.ShowMemoryMap = executor.ShowMemoryMap
	runtime.Executor.SetMode = executor.SetMode
	runtime.ExecutorConfigured = true
	runtime.plannerMode = PlannerModeDeterministic
	return runtime
}

func (r Runtime) PlannerMode() PlannerMode {
	return r.plannerMode
}

func (r Runtime) CurrentPlanner() ActionResult {
	if r.plannerMode == PlannerModeLLM {
		return ActionResult{OK: true, Message: MessageLLMMode}
	}
	return ActionResult{OK: true, Message: MessageDeterministicMode}
}

func (r *Runtime) SetPlannerMode(mode PlannerMode) ActionResult {
	if r == nil {
		return ActionResult{OK: false, Message: MessagePlannerMissing}
	}
	switch mode {
	case PlannerModeDeterministic:
		r.plannerMode = PlannerModeDeterministic
		return ActionResult{OK: true, Message: MessagePlannerSwitchedDeterministic}
	case PlannerModeLLM:
		if r.llmPlanner == nil || !r.llmPlanner.Available() {
			return ActionResult{OK: false, Message: MessageLLMModeNotConfigured}
		}
		r.plannerMode = PlannerModeLLM
		return ActionResult{OK: true, Message: MessagePlannerSwitchedLLM}
	default:
		return ActionResult{OK: false, Message: MessageUnsupportedMode}
	}
}

func (r *Runtime) ConfigureLLMPlanner(planner Planner) {
	if r == nil {
		return
	}
	if planner == nil || !planner.Available() {
		r.llmPlanner = nil
		if r.plannerMode == PlannerModeLLM {
			r.plannerMode = PlannerModeDeterministic
		}
		return
	}
	r.llmPlanner = planner
}

func (r *Runtime) CopyPlannerConfiguration(source *Runtime) {
	if r == nil {
		return
	}
	if source == nil {
		r.plannerMode = PlannerModeDeterministic
		r.llmPlanner = nil
		return
	}
	r.plannerMode = source.plannerMode
	r.llmPlanner = source.llmPlanner
}

func (r Runtime) RunAction(kind ActionKind, intent IntentKind, risk RiskLevel, target *[MaxNameLen]byte, targetLen int, context *Context) Response {
	if context != nil {
		context.BeginRequest(nil, 0, r.plannerMode)
	}
	return r.runAction(kind, intent, risk, target, targetLen, context)
}

func (r Runtime) RunActionRequest(kind ActionKind, intent IntentKind, risk RiskLevel, target *[MaxNameLen]byte, targetLen int, input *[MaxContextInput]byte, inputLen int, context *Context) Response {
	if context != nil {
		context.BeginRequest(input, inputLen, r.plannerMode)
	}
	return r.runAction(kind, intent, risk, target, targetLen, context)
}

func (r Runtime) runAction(kind ActionKind, intent IntentKind, risk RiskLevel, target *[MaxNameLen]byte, targetLen int, context *Context) Response {
	plan := singleActionPlan(r.plannerMode, intent, kind, risk)
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
	if r == nil {
		return MessageExecutorNotConfigured
	}
	if targetLen < 0 || targetLen > MaxNameLen {
		return MessageActionTargetInvalid
	}
	return r.RunAction(kind, intent, risk, target, targetLen, context).Result.Message
}

func (r *Runtime) RunActionRequestMessage(kind ActionKind, intent IntentKind, risk RiskLevel, target *[MaxNameLen]byte, targetLen int, input *[MaxContextInput]byte, inputLen int, context *Context) MessageKind {
	if r == nil {
		return MessageExecutorNotConfigured
	}
	if targetLen < 0 || targetLen > MaxNameLen {
		return MessageActionTargetInvalid
	}
	return r.RunActionRequest(kind, intent, risk, target, targetLen, input, inputLen, context).Result.Message
}

func (r *Runtime) ConfirmActionMessage(confirmed bool, context *Context) MessageKind {
	if r == nil {
		return MessageExecutorNotConfigured
	}
	return r.ConfirmAction(confirmed, context).Result.Message
}

func (r Runtime) ConfirmAction(confirmed bool, context *Context) Response {
	var response Response
	if context == nil || !context.ConfirmationPending {
		setResponseResult(&response, false, MessageActionCancelled)
		setSafety(&response, SafetyRejected, MessageActionCancelled)
		return response
	}

	plan := context.PendingPlan
	context.PendingPlan = Plan{}
	context.ConfirmationPending = false
	if !confirmed {
		setResponseResult(&response, false, MessageActionCancelled)
		setSafety(&response, SafetyRejected, MessageActionCancelled)
		response.AddTrace(TraceSafety, TraceDetailRejected)
		return finishPlanContext(response, context, plan)
	}
	return r.runPlanWithConfirmation(plan, context, true)
}

func (r Runtime) runPlan(plan Plan, context *Context) Response {
	return r.runPlanWithConfirmation(plan, context, false)
}

func (r Runtime) runPlanWithConfirmation(plan Plan, context *Context, confirmed bool) Response {
	var response Response
	response.AddTrace(TracePlanner, plannerTrace(plan.Planner))
	response.AddTrace(TraceIntent, intentTrace(plan.Intent))

	validation := validatePlan(plan)
	if !validation.OK {
		setResponseResult(&response, false, validation.Reason)
		setSafety(&response, SafetyRejected, MessageValidationFailed)
		response.AddTrace(TraceValidation, traceFromMessage(validation.Reason))
		return finishPlanContext(response, context, plan)
	}
	response.AddTrace(TraceValidation, TraceDetailOK)
	if context != nil {
		context.CurrentTask = plan.Actions[0].Kind
	}

	safety := evaluateSafetyWithConfirmation(plan, context, confirmed)
	setSafety(&response, safety.Status, safety.Reason)
	response.AddTrace(TraceSafety, safetyTrace(safety.Status))
	if safety.Status != SafetyAllowed {
		setResponseResult(&response, false, safety.Reason)
		return finishPlanContext(response, context, plan)
	}

	if !r.ExecutorConfigured {
		setResponseResult(&response, false, MessageExecutorNotConfigured)
		response.AddTrace(TraceExecutor, TraceDetailMissing)
		return finishPlanContext(response, context, plan)
	}

	var results [MaxActions]ActionResult
	for i := 0; i < plan.ActionCount; i++ {
		if context != nil {
			context.CurrentTask = plan.Actions[i].Kind
		}
		results[i] = r.Executor.Execute(plan.Actions[i], context)
		if !results[i].OK {
			response.AddTrace(TraceExecutor, TraceDetailFailed)
			setResponseResult(&response, results[i].OK, results[i].Message)
			return finishPlanContext(response, context, plan)
		}
	}
	response.AddTrace(TraceExecutor, TraceDetailSuccess)

	result := formatResult(plan, results, plan.ActionCount, safety)
	setResponseResult(&response, result.OK, result.Message)
	response.AddTrace(TraceFormatter, TraceDetailStructured)
	return finishPlanContext(response, context, plan)
}

func finishPlanContext(response Response, context *Context, plan Plan) Response {
	if context == nil {
		return response
	}
	context.LastIntent = plan.Intent
	context.LastAction = ActionNone
	if plan.ActionCount > 0 {
		actionIndex := plan.ActionCount - 1
		if actionIndex >= MaxActions {
			actionIndex = MaxActions - 1
		}
		context.LastAction = plan.Actions[actionIndex].Kind
	}
	context.LastResultSummary = response.Result.Message
	if context.LastAction != ActionSetMode {
		context.PlannerMode = plan.Planner
	}
	if response.Safety.Status != SafetyConfirmationRequired {
		context.CurrentTask = ActionNone
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
		if !action.Kind.Valid() || plan.Planner == PlannerModeLLM && !allowedPlanAction(action.Kind) {
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
		if action.DataLen != 0 {
			return ValidationResult{OK: false, Reason: MessageActionDataInvalid}
		}
		if actionRequiresTarget(action.Kind) && action.TargetLen == 0 {
			return ValidationResult{OK: false, Reason: MessageActionTargetInvalid}
		}
		if plan.Planner == PlannerModeLLM && actionRisk(action.Kind) != action.Risk {
			return ValidationResult{OK: false, Reason: MessageActionRiskInvalid}
		}
	}
	return ValidationResult{OK: true, Reason: MessageOK}
}

func allowedPlanAction(kind ActionKind) bool {
	switch kind {
	case ActionListFiles, ActionReadFile, ActionDeleteFile, ActionStatFile, ActionShowHelp, ActionShowHistory, ActionShowVersion, ActionShowTicks, ActionShowMemoryMap, ActionSetMode:
		return true
	default:
		return false
	}
}

func actionRequiresTarget(kind ActionKind) bool {
	switch kind {
	case ActionReadFile, ActionDeleteFile, ActionStatFile:
		return true
	default:
		return false
	}
}

func actionRisk(kind ActionKind) RiskLevel {
	if kind == ActionDeleteFile {
		return RiskRisky
	}
	return RiskSafe
}

func validRiskLevel(risk RiskLevel) bool {
	switch risk {
	case RiskSafe, RiskRisky:
		return true
	default:
		return false
	}
}

func evaluateSafety(plan Plan, context *Context) SafetyDecision {
	return evaluateSafetyWithConfirmation(plan, context, false)
}

func evaluateSafetyWithConfirmation(plan Plan, context *Context, confirmed bool) SafetyDecision {
	for i := 0; i < plan.ActionCount; i++ {
		if plan.Actions[i].Risk == RiskRisky {
			if confirmed {
				return SafetyDecision{Status: SafetyAllowed, Reason: MessageOK}
			}
			if context != nil {
				context.PendingPlan = plan
				context.ConfirmationPending = true
			}
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
	case IntentShowVersion:
		return TraceDetailShowVersion
	case IntentShowTicks:
		return TraceDetailShowTicks
	case IntentShowMemoryMap:
		return TraceDetailShowMemoryMap
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
