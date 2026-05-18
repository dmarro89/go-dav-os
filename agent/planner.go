package agent

func singleActionPlan(mode PlannerMode, intent IntentKind, actionKind ActionKind, risk RiskLevel) Plan {
	var plan Plan
	plan.Planner = mode
	plan.Intent = intent
	plan.Confidence = 100
	plan.ActionCount = 1
	plan.Actions[0].Kind = actionKind
	plan.Actions[0].Risk = risk
	return plan
}
