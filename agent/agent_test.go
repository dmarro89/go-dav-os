package agent

import "testing"

func TestRuntimeExecutesTypedSafePlan(t *testing.T) {
	executed := false
	runtime := Runtime{
		Planner: DeterministicPlanner{},
		Executor: AllowedActionExecutor{
			ListFiles: func(action Action, context *Context) ActionResult {
				executed = true
				if action.Kind != ActionListFiles {
					t.Fatalf("unexpected action kind: %v", action.Kind)
				}
				return ActionResult{OK: true, Message: "files listed"}
			},
		},
	}

	var context Context
	response := runtime.Run("show files", &context)
	if !response.Result.OK {
		t.Fatalf("expected successful response, got %q", response.Result.Message)
	}
	if !executed {
		t.Fatalf("expected executor to run")
	}
	if response.Safety.Status != SafetyAllowed {
		t.Fatalf("expected safety allowed, got %v", response.Safety.Status)
	}
	if context.LastIntent != IntentListFiles {
		t.Fatalf("expected context to remember list intent, got %v", context.LastIntent)
	}
}

func TestRuntimeRequiresConfirmationForRiskyPlan(t *testing.T) {
	executed := false
	runtime := Runtime{
		Planner: staticPlanner{plan: singleActionPlan(PlannerModeLLM, IntentDeleteFile, ActionDeleteFile, RiskRisky)},
		Executor: AllowedActionExecutor{
			DeleteFile: func(action Action, context *Context) ActionResult {
				executed = true
				return ActionResult{OK: true, Message: "deleted"}
			},
		},
	}

	response := runtime.Run("delete notes", nil)
	if response.Safety.Status != SafetyConfirmationRequired {
		t.Fatalf("expected confirmation_required, got %v", response.Safety.Status)
	}
	if response.Result.OK {
		t.Fatalf("expected response to stop before execution")
	}
	if executed {
		t.Fatalf("risky action executed without confirmation")
	}
}

func TestValidatorRejectsUnsupportedActionKinds(t *testing.T) {
	plan := singleActionPlan(PlannerModeLLM, IntentUnknown, ActionKind(99), RiskSafe)
	result := DefaultValidator{}.Validate(plan)
	if result.OK {
		t.Fatalf("expected unsupported action to be rejected")
	}
}

func TestValidatorRejectsUnsupportedRiskLevels(t *testing.T) {
	plan := singleActionPlan(PlannerModeLLM, IntentDeleteFile, ActionDeleteFile, RiskLevel(99))
	result := DefaultValidator{}.Validate(plan)
	if result.OK {
		t.Fatalf("expected unsupported risk to be rejected")
	}
}

func TestLLMPlannerDelegatesToBridge(t *testing.T) {
	bridge := fakeBridge{plan: singleActionPlan(PlannerModeDeterministic, IntentStatFile, ActionStatFile, RiskSafe)}
	result := LLMPlanner{Bridge: bridge}.Plan("status of notes", nil)
	if !result.OK {
		t.Fatalf("expected LLM planner to succeed, got %q", result.Reason)
	}
	plan := result.Plan
	if plan.Planner != PlannerModeLLM {
		t.Fatalf("expected LLM planner mode, got %v", plan.Planner)
	}
	if plan.Intent != IntentStatFile {
		t.Fatalf("expected bridge intent, got %v", plan.Intent)
	}
}

func TestLLMPlannerFailsWithoutBridge(t *testing.T) {
	result := LLMPlanner{}.Plan("show files", nil)
	if result.OK {
		t.Fatalf("expected missing bridge to fail")
	}
	if result.Reason != errLLMBridgeNotConfigured {
		t.Fatalf("unexpected failure reason: %q", result.Reason)
	}
}

func TestRuntimeReturnsPlannerError(t *testing.T) {
	response := Runtime{Planner: failingPlanner{reason: errLLMBridgeNotConfigured}}.Run("show files", nil)
	if response.Result.OK {
		t.Fatalf("expected planner failure response")
	}
	if response.Result.Message != errLLMBridgeNotConfigured {
		t.Fatalf("unexpected planner failure message: %q", response.Result.Message)
	}
	if response.Safety.Status != SafetyRejected {
		t.Fatalf("expected rejected safety status, got %v", response.Safety.Status)
	}
}

func TestRuntimeReturnsDefaultPlannerError(t *testing.T) {
	response := Runtime{Planner: failingPlanner{}}.Run("show files", nil)
	if response.Result.Message != "agent: planner failed" {
		t.Fatalf("unexpected planner failure message: %q", response.Result.Message)
	}
	if response.TraceCount != 1 || response.Trace[0].Stage != "Planner" {
		t.Fatalf("expected planner trace, got count=%d", response.TraceCount)
	}
}

func TestRuntimeRejectsMissingPlanner(t *testing.T) {
	response := Runtime{}.Run("show files", nil)
	if response.Result.OK {
		t.Fatalf("expected missing planner to fail")
	}
	if response.Safety.Reason != "planner_missing" {
		t.Fatalf("unexpected safety reason: %q", response.Safety.Reason)
	}
}

func TestRuntimeStopsOnValidationFailure(t *testing.T) {
	executed := false
	response := Runtime{
		Planner: staticPlanner{plan: Plan{}},
		Executor: AllowedActionExecutor{
			ShowHelp: func(action Action, context *Context) ActionResult {
				executed = true
				return ActionResult{OK: true, Message: "help"}
			},
		},
	}.Run("invalid", nil)

	if response.Result.OK {
		t.Fatalf("expected validation failure")
	}
	if response.Safety.Reason != "validation_failed" {
		t.Fatalf("unexpected safety reason: %q", response.Safety.Reason)
	}
	if executed {
		t.Fatalf("executor ran after validation failure")
	}
}

func TestRuntimeReportsMissingExecutor(t *testing.T) {
	response := Runtime{
		Planner: staticPlanner{plan: singleActionPlan(PlannerModeDeterministic, IntentShowHelp, ActionShowHelp, RiskSafe)},
	}.Run("help", nil)

	if response.Result.Message != "agent: executor not configured" {
		t.Fatalf("unexpected missing executor message: %q", response.Result.Message)
	}
}

func TestRuntimeStopsOnExecutorFailure(t *testing.T) {
	response := Runtime{
		Planner: staticPlanner{plan: singleActionPlan(PlannerModeDeterministic, IntentReadFile, ActionReadFile, RiskSafe)},
		Executor: AllowedActionExecutor{
			ReadFile: func(action Action, context *Context) ActionResult {
				return ActionResult{OK: false, Message: "read failed"}
			},
		},
	}.Run("read notes", nil)

	if response.Result.OK {
		t.Fatalf("expected executor failure")
	}
	if response.Result.Message != "read failed" {
		t.Fatalf("unexpected executor failure: %q", response.Result.Message)
	}
}

func TestAllowedActionExecutorFailsClosed(t *testing.T) {
	result := AllowedActionExecutor{}.Execute(Action{Kind: ActionReadFile}, nil)
	if result.OK {
		t.Fatalf("expected unavailable action to fail")
	}
}

func TestAllowedActionExecutorDispatchesAllActions(t *testing.T) {
	calls := 0
	handler := func(expected ActionKind) ActionHandler {
		return func(action Action, context *Context) ActionResult {
			calls++
			if action.Kind != expected {
				t.Fatalf("expected action %v, got %v", expected, action.Kind)
			}
			return ActionResult{OK: true, Message: action.Kind.String()}
		}
	}

	executor := AllowedActionExecutor{
		ListFiles:  handler(ActionListFiles),
		ReadFile:   handler(ActionReadFile),
		WriteFile:  handler(ActionWriteFile),
		DeleteFile: handler(ActionDeleteFile),
		StatFile:   handler(ActionStatFile),
		ShowHelp:   handler(ActionShowHelp),
	}

	actions := [...]ActionKind{
		ActionListFiles,
		ActionReadFile,
		ActionWriteFile,
		ActionDeleteFile,
		ActionStatFile,
		ActionShowHelp,
	}
	for _, kind := range actions {
		result := executor.Execute(Action{Kind: kind}, nil)
		if !result.OK {
			t.Fatalf("expected action %v to succeed: %q", kind, result.Message)
		}
	}
	if calls != len(actions) {
		t.Fatalf("expected %d calls, got %d", len(actions), calls)
	}
}

func TestAllowedActionExecutorRejectsUnknownAction(t *testing.T) {
	result := AllowedActionExecutor{}.Execute(Action{Kind: ActionKind(99)}, nil)
	if result.OK {
		t.Fatalf("expected unknown action to fail")
	}
	if result.Message != "agent: unsupported action" {
		t.Fatalf("unexpected result: %q", result.Message)
	}
}

func TestDefaultValidatorRejectsMalformedPlans(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
	}{
		{name: "no actions", plan: Plan{}},
		{name: "too many actions", plan: Plan{ActionCount: MaxActions + 1}},
		{name: "target too long", plan: planWithAction(Action{Kind: ActionReadFile, Risk: RiskSafe, TargetLen: MaxNameLen + 1})},
		{name: "negative target", plan: planWithAction(Action{Kind: ActionReadFile, Risk: RiskSafe, TargetLen: -1})},
		{name: "data too long", plan: planWithAction(Action{Kind: ActionWriteFile, Risk: RiskSafe, DataLen: MaxDataLen + 1})},
		{name: "negative data", plan: planWithAction(Action{Kind: ActionWriteFile, Risk: RiskSafe, DataLen: -1})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultValidator{}.Validate(tt.plan)
			if result.OK {
				t.Fatalf("expected malformed plan to be rejected")
			}
		})
	}
}

func TestDefaultFormatter(t *testing.T) {
	var results [MaxActions]ActionResult
	formatter := DefaultFormatter{}

	if result := formatter.Format(Plan{}, results, 0, SafetyDecision{}); result.OK {
		t.Fatalf("expected no-result format to fail")
	}

	results[0] = ActionResult{OK: true, Message: "one"}
	if result := formatter.Format(Plan{}, results, 1, SafetyDecision{}); result.Message != "one" {
		t.Fatalf("expected single result message, got %q", result.Message)
	}

	results[1] = ActionResult{OK: true, Message: "two"}
	if result := formatter.Format(Plan{}, results, 2, SafetyDecision{}); !result.OK || result.Message != "agent: completed plan" {
		t.Fatalf("unexpected multi result: %+v", result)
	}
}

func TestDeterministicPlannerRecognizesHelpAndDefaultsSafely(t *testing.T) {
	tests := []struct {
		input  string
		intent IntentKind
		action ActionKind
	}{
		{input: "HELP", intent: IntentShowHelp, action: ActionShowHelp},
		{input: "what can you do?", intent: IntentShowHelp, action: ActionShowHelp},
		{input: "help with files", intent: IntentShowHelp, action: ActionShowHelp},
		{input: "LS", intent: IntentListFiles, action: ActionListFiles},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := DeterministicPlanner{}.Plan(tt.input, nil)
			if !result.OK {
				t.Fatalf("expected deterministic plan to succeed")
			}
			if result.Plan.Intent != tt.intent || result.Plan.Actions[0].Kind != tt.action {
				t.Fatalf("unexpected plan: intent=%v action=%v", result.Plan.Intent, result.Plan.Actions[0].Kind)
			}
		})
	}
}

func TestLLMPlannerPropagatesBridgeFailure(t *testing.T) {
	result := LLMPlanner{Bridge: failingBridge{reason: "bridge timeout"}}.Plan("show files", nil)
	if result.OK {
		t.Fatalf("expected bridge failure")
	}
	if result.Reason != "bridge timeout" {
		t.Fatalf("unexpected bridge failure: %q", result.Reason)
	}
}

func TestLLMPlannerDefaultsEmptyBridgeFailureReason(t *testing.T) {
	result := LLMPlanner{Bridge: failingBridge{}}.Plan("show files", nil)
	if result.Reason != "agent: llm bridge failed" {
		t.Fatalf("unexpected bridge failure: %q", result.Reason)
	}
}

func TestResponseAddTraceStopsAtCapacity(t *testing.T) {
	var response Response
	for i := 0; i < MaxTraceEntries+2; i++ {
		response.AddTrace("stage", "detail")
	}
	if response.TraceCount != MaxTraceEntries {
		t.Fatalf("expected trace count %d, got %d", MaxTraceEntries, response.TraceCount)
	}
}

func TestContextRememberRollsRecentItems(t *testing.T) {
	var context Context
	for i := 0; i < MaxRecentItems+1; i++ {
		context.Remember("input", "result", IntentShowHelp)
	}
	if context.RecentCount != MaxRecentItems {
		t.Fatalf("expected recent count %d, got %d", MaxRecentItems, context.RecentCount)
	}
	if context.LastIntent != IntentShowHelp {
		t.Fatalf("unexpected last intent: %v", context.LastIntent)
	}
}

func TestEnumStrings(t *testing.T) {
	if PlannerModeDeterministic.String() != "deterministic" || PlannerModeLLM.String() != "llm" || PlannerMode(99).String() != stringUnknown {
		t.Fatalf("unexpected planner mode strings")
	}
	if IntentReadFile.String() != stringReadFile || IntentWriteFile.String() != stringWriteFile || IntentDeleteFile.String() != stringDeleteFile || IntentKind(99).String() != stringUnknown {
		t.Fatalf("unexpected intent strings")
	}
	if ActionNone.String() != stringNone || ActionDeleteFile.String() != stringDeleteFile || ActionKind(99).String() != stringNone {
		t.Fatalf("unexpected action strings")
	}
	if RiskSafe.String() != "safe" || RiskRisky.String() != "risky" {
		t.Fatalf("unexpected risk strings")
	}
	if SafetyAllowed.String() != "allowed" || SafetyConfirmationRequired.String() != "confirmation_required" || SafetyRejected.String() != "rejected" {
		t.Fatalf("unexpected safety strings")
	}
}

type staticPlanner struct {
	plan Plan
}

func (p staticPlanner) Plan(input string, context *Context) PlanningResult {
	return successfulPlan(p.plan)
}

type failingPlanner struct {
	reason string
}

func (p failingPlanner) Plan(input string, context *Context) PlanningResult {
	return PlanningResult{OK: false, Reason: p.reason}
}

type fakeBridge struct {
	plan Plan
}

func (b fakeBridge) Plan(input string, context *Context) PlanningResult {
	return successfulPlan(b.plan)
}

type failingBridge struct {
	reason string
}

func (b failingBridge) Plan(input string, context *Context) PlanningResult {
	return PlanningResult{OK: false, Reason: b.reason}
}

func planWithAction(action Action) Plan {
	var plan Plan
	plan.ActionCount = 1
	plan.Actions[0] = action
	return plan
}
