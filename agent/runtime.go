package agent

type Planner interface {
	Plan(input string, context *Context) PlanningResult
}

type Validator interface {
	Validate(plan Plan) ValidationResult
}

type SafetyGate interface {
	Evaluate(plan Plan, context *Context) SafetyDecision
}

type Executor interface {
	Execute(action Action, context *Context) ActionResult
}

type Formatter interface {
	Format(plan Plan, results [MaxActions]ActionResult, resultCount int, safety SafetyDecision) ActionResult
}

type Runtime struct {
	Planner    Planner
	Validator  Validator
	SafetyGate SafetyGate
	Executor   Executor
	Formatter  Formatter
}

// Run executes the shared agent pipeline:
// planner -> validator -> safety gate -> typed action executor -> formatter
func (r Runtime) Run(input string, context *Context) Response {
	var response Response
	if r.Planner == nil {
		response.Result = ActionResult{OK: false, Message: "agent: planner not configured"}
		response.Safety = SafetyDecision{Status: SafetyRejected, Reason: "planner_missing"}
		response.AddTrace("Planner", "missing")
		return response
	}

	planning := r.Planner.Plan(input, context)
	if !planning.OK {
		reason := planning.Reason
		if reason == "" {
			reason = "agent: planner failed"
		}
		response.Result = ActionResult{OK: false, Message: reason}
		response.Safety = SafetyDecision{Status: SafetyRejected, Reason: "planner_failed"}
		response.AddTrace("Planner", reason)
		return response
	}

	plan := planning.Plan
	response.AddTrace("Planner", plan.Planner.String())
	response.AddTrace("Intent", plan.Intent.String())

	validator := r.Validator
	if validator == nil {
		validator = DefaultValidator{}
	}
	validation := validator.Validate(plan)
	if !validation.OK {
		response.Result = ActionResult{OK: false, Message: validation.Reason}
		response.Safety = SafetyDecision{Status: SafetyRejected, Reason: "validation_failed"}
		response.AddTrace("Validation", validation.Reason)
		return response
	}
	response.AddTrace("Validation", "ok")

	gate := r.SafetyGate
	if gate == nil {
		gate = DefaultSafetyGate{}
	}
	safety := gate.Evaluate(plan, context)
	response.Safety = safety
	response.AddTrace("Safety", safety.Status.String())
	if safety.Status != SafetyAllowed {
		response.Result = ActionResult{OK: false, Message: safety.Reason}
		return response
	}

	if r.Executor == nil {
		response.Result = ActionResult{OK: false, Message: "agent: executor not configured"}
		response.AddTrace("Executor", "missing")
		return response
	}

	var results [MaxActions]ActionResult
	for i := 0; i < plan.ActionCount; i++ {
		results[i] = r.Executor.Execute(plan.Actions[i], context)
		if !results[i].OK {
			response.AddTrace("Executor", "failed")
			response.Result = results[i]
			return response
		}
	}
	response.AddTrace("Executor", "success")

	formatter := r.Formatter
	if formatter == nil {
		formatter = DefaultFormatter{}
	}
	response.Result = formatter.Format(plan, results, plan.ActionCount, safety)
	response.AddTrace("Formatter", "structured")
	if context != nil {
		context.Remember(input, response.Result.Message, plan.Intent)
	}
	return response
}

type DefaultValidator struct{}

func (DefaultValidator) Validate(plan Plan) ValidationResult {
	if plan.ActionCount <= 0 {
		return ValidationResult{OK: false, Reason: "agent: plan has no actions"}
	}
	if plan.ActionCount > MaxActions {
		return ValidationResult{OK: false, Reason: "agent: plan has too many actions"}
	}
	for i := 0; i < plan.ActionCount; i++ {
		action := plan.Actions[i]
		if !validActionKind(action.Kind) {
			return ValidationResult{OK: false, Reason: "agent: plan contains unsupported action"}
		}
		if !validRiskLevel(action.Risk) {
			return ValidationResult{OK: false, Reason: "agent: action risk is invalid"}
		}
		if action.TargetLen < 0 || action.TargetLen > MaxNameLen {
			return ValidationResult{OK: false, Reason: "agent: action target is invalid"}
		}
		if action.DataLen < 0 || action.DataLen > MaxDataLen {
			return ValidationResult{OK: false, Reason: "agent: action data is invalid"}
		}
	}
	return ValidationResult{OK: true, Reason: "ok"}
}

func validActionKind(kind ActionKind) bool {
	switch kind {
	case ActionListFiles, ActionReadFile, ActionWriteFile, ActionDeleteFile, ActionStatFile, ActionShowHelp:
		return true
	default:
		return false
	}
}

func validRiskLevel(risk RiskLevel) bool {
	switch risk {
	case RiskSafe, RiskRisky:
		return true
	default:
		return false
	}
}

type DefaultSafetyGate struct{}

func (DefaultSafetyGate) Evaluate(plan Plan, _ *Context) SafetyDecision {
	for i := 0; i < plan.ActionCount; i++ {
		if plan.Actions[i].Risk == RiskRisky {
			return SafetyDecision{Status: SafetyConfirmationRequired, Reason: "agent: confirmation required"}
		}
	}
	return SafetyDecision{Status: SafetyAllowed, Reason: "ok"}
}

type DefaultFormatter struct{}

func (DefaultFormatter) Format(_ Plan, results [MaxActions]ActionResult, resultCount int, _ SafetyDecision) ActionResult {
	if resultCount <= 0 {
		return ActionResult{OK: false, Message: "agent: no result"}
	}
	if resultCount == 1 {
		return results[0]
	}
	return ActionResult{OK: true, Message: "agent: completed plan"}
}
