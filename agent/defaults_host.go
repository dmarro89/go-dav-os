//go:build !gccgo

package agent

type DefaultValidator struct{}

func (DefaultValidator) Validate(plan Plan) ValidationResult {
	return validatePlan(plan)
}

type DefaultSafetyGate struct{}

func (DefaultSafetyGate) Evaluate(plan Plan, context *Context) SafetyDecision {
	return evaluateSafety(plan, context)
}

type DefaultFormatter struct{}

func (DefaultFormatter) Format(plan Plan, results [MaxActions]ActionResult, resultCount int, safety SafetyDecision) ActionResult {
	return formatResult(plan, results, resultCount, safety)
}
