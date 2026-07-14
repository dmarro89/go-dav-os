package agent

import "testing"

func TestRuntimeExecutesTypedSafePlan(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(
		AllowedActionExecutor{
			ListFiles: func(action Action, context *Context) ActionResult {
				executed = true
				if action.Kind != ActionListFiles {
					t.Fatalf("unexpected action kind: %v", action.Kind)
				}
				return ActionResult{OK: true, Message: MessageFilesListed}
			},
		},
	)

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
	runtime := NewDeterministicAgent(
		AllowedActionExecutor{
			DeleteFile: func(action Action, context *Context) ActionResult {
				executed = true
				return ActionResult{OK: true, Message: MessageOK}
			},
		},
	)

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
	tests := [...]ActionKind{ActionUnknown, ActionKind(99)}

	for _, kind := range tests {
		plan := singleActionPlan(PlannerModeLLM, IntentUnknown, kind, RiskSafe)
		result := DefaultValidator{}.Validate(plan)
		if result.OK {
			t.Fatalf("expected unsupported action %v to be rejected", kind)
		}
	}
}

func TestIssue153TypedPlanContract(t *testing.T) {
	plan := singleActionPlan(PlannerModeLLM, IntentShowVersion, ActionShowVersion, RiskSafe)

	if plan.Intent != IntentShowVersion || plan.Actions[0].Kind != ActionShowVersion {
		t.Fatalf("typed plan did not preserve intent/action: %+v", plan)
	}
	if !plan.Actions[0].Kind.Valid() {
		t.Fatalf("expected issue #153 action to be known")
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
	if result.Reason != MessageLLMBridgeNotConfigured {
		t.Fatalf("unexpected failure reason: %q", result.Reason)
	}
}

func TestRuntimeStopsOnValidationFailure(t *testing.T) {
	executed := false
	response := NewDeterministicAgent(
		AllowedActionExecutor{
			ShowHelp: func(action Action, context *Context) ActionResult {
				executed = true
				return ActionResult{OK: true, Message: MessageHelp}
			},
		},
	).runPlan(Plan{}, nil)

	if response.Result.OK {
		t.Fatalf("expected validation failure")
	}
	if response.Safety.Reason != MessageValidationFailed {
		t.Fatalf("unexpected safety reason: %q", response.Safety.Reason)
	}
	if executed {
		t.Fatalf("executor ran after validation failure")
	}
}

func TestRuntimeReportsMissingExecutor(t *testing.T) {
	response := Runtime{}.RunAction(ActionShowHelp, IntentShowHelp, RiskSafe, nil, 0, nil)

	if response.Result.Message != MessageExecutorNotConfigured {
		t.Fatalf("unexpected missing executor message: %q", response.Result.Message)
	}
}

func TestRunActionExecutesTypedWriteAction(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		WriteFile: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageOK}
		},
	})

	response := runtime.RunAction(ActionWriteFile, IntentWriteFile, RiskSafe, nil, 0, nil)
	if !response.Result.OK || !executed {
		t.Fatalf("expected typed write action to execute, got %+v", response.Result)
	}
}

func TestRunActionExecutesConfirmedDeleteAction(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		DeleteFile: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageOK}
		},
	})
	target := [MaxNameLen]byte{'n', 'o', 't', 'e', 's'}

	response := runtime.RunAction(ActionDeleteFile, IntentDeleteFile, RiskSafe, &target, 5, nil)
	if !response.Result.OK || !executed {
		t.Fatalf("expected confirmed delete action to execute, got %+v", response.Result)
	}
}

func TestRuntimeStopsOnExecutorFailure(t *testing.T) {
	response := NewDeterministicAgent(
		AllowedActionExecutor{
			ReadFile: func(action Action, context *Context) ActionResult {
				return ActionResult{OK: false, Message: MessageReadFailed}
			},
		},
	).Run("read notes", nil)

	if response.Result.OK {
		t.Fatalf("expected executor failure")
	}
	if response.Result.Message != MessageReadFailed {
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
			return ActionResult{OK: true, Message: MessageOK}
		}
	}

	executor := AllowedActionExecutor{
		ListFiles:     handler(ActionListFiles),
		ReadFile:      handler(ActionReadFile),
		WriteFile:     handler(ActionWriteFile),
		DeleteFile:    handler(ActionDeleteFile),
		StatFile:      handler(ActionStatFile),
		ShowHelp:      handler(ActionShowHelp),
		ShowHistory:   handler(ActionShowHistory),
		ShowVersion:   handler(ActionShowVersion),
		ShowTicks:     handler(ActionShowTicks),
		ShowMemoryMap: handler(ActionShowMemoryMap),
		SetMode:       handler(ActionSetMode),
	}

	actions := [...]ActionKind{
		ActionListFiles,
		ActionReadFile,
		ActionWriteFile,
		ActionDeleteFile,
		ActionStatFile,
		ActionShowHelp,
		ActionShowHistory,
		ActionShowVersion,
		ActionShowTicks,
		ActionShowMemoryMap,
		ActionSetMode,
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

func TestNewDeterministicAgentUsesDeterministicPlanner(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		ShowHelp: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageHelp}
		},
	})

	response := runtime.Run("help", nil)
	if !response.Result.OK || response.Result.Message != MessageHelp {
		t.Fatalf("unexpected deterministic agent response: %+v", response.Result)
	}
	if !executed {
		t.Fatalf("expected configured executor to run")
	}
	if response.TraceCount == 0 || response.Trace[0].Detail != TraceDetailDeterministic {
		t.Fatalf("expected deterministic planner trace, got %+v", response.Trace[0])
	}
}

func TestAllowedActionExecutorRejectsUnknownAction(t *testing.T) {
	result := AllowedActionExecutor{}.Execute(Action{Kind: ActionUnknown}, nil)
	if result.OK {
		t.Fatalf("expected unknown action to fail")
	}
	if result.Message != MessageUnsupportedAction {
		t.Fatalf("unexpected result: %q", result.Message)
	}
}

func TestKnownButUnwiredActionsFailClosed(t *testing.T) {
	result := AllowedActionExecutor{}.Execute(Action{Kind: ActionShowTicks}, nil)
	if result.OK {
		t.Fatalf("expected unwired action to fail")
	}
	if result.Message != MessageActionUnavailable {
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
		{name: "data too long", plan: planWithAction(Action{Kind: ActionReadFile, Risk: RiskSafe, TargetLen: 1, DataLen: MaxDataLen + 1})},
		{name: "negative data", plan: planWithAction(Action{Kind: ActionReadFile, Risk: RiskSafe, TargetLen: 1, DataLen: -1})},
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

func TestDefaultValidatorRejectsUnsafePlanShapes(t *testing.T) {
	tests := []struct {
		name   string
		plan   Plan
		reason MessageKind
	}{
		{
			name:   "action outside allowlist",
			plan:   llmPlanWithAction(Action{Kind: ActionWriteFile, Risk: RiskSafe}),
			reason: MessagePlanContainsUnsupportedAction,
		},
		{
			name:   "missing read target",
			plan:   llmPlanWithAction(Action{Kind: ActionReadFile, Risk: RiskSafe}),
			reason: MessageActionTargetInvalid,
		},
		{
			name:   "missing delete target",
			plan:   llmPlanWithAction(Action{Kind: ActionDeleteFile, Risk: RiskRisky}),
			reason: MessageActionTargetInvalid,
		},
		{
			name:   "delete marked safe",
			plan:   llmPlanWithTargetAction(ActionDeleteFile, RiskSafe, "notes"),
			reason: MessageActionRiskInvalid,
		},
		{
			name:   "read marked risky",
			plan:   llmPlanWithTargetAction(ActionReadFile, RiskRisky, "notes"),
			reason: MessageActionRiskInvalid,
		},
		{
			name: "raw action data",
			plan: llmPlanWithAction(Action{
				Kind:      ActionReadFile,
				Risk:      RiskSafe,
				TargetLen: 1,
				DataLen:   1,
			}),
			reason: MessageActionDataInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultValidator{}.Validate(tt.plan)
			if result.OK {
				t.Fatalf("expected invalid plan to be rejected")
			}
			if result.Reason != tt.reason {
				t.Fatalf("validation reason = %q, expected %q", result.Reason, tt.reason)
			}
		})
	}
}

func TestDefaultValidatorAllowsSafeNoTargetActions(t *testing.T) {
	actions := [...]ActionKind{
		ActionListFiles,
		ActionShowHelp,
		ActionShowHistory,
		ActionShowVersion,
		ActionShowTicks,
		ActionShowMemoryMap,
		ActionSetMode,
	}

	for _, kind := range actions {
		plan := planWithAction(Action{Kind: kind, Risk: RiskSafe})
		if result := (DefaultValidator{}).Validate(plan); !result.OK {
			t.Fatalf("expected action %v to validate, got %q", kind, result.Reason)
		}
	}
}

func TestRuntimeRejectsInvalidLLMPlanBeforeExecution(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		ReadFile: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageFileRead}
		},
	})

	response := runtime.runPlan(singleActionPlan(PlannerModeLLM, IntentReadFile, ActionReadFile, RiskSafe), nil)
	if response.Result.OK {
		t.Fatalf("expected invalid LLM plan to fail")
	}
	if response.Result.Message != MessageActionTargetInvalid {
		t.Fatalf("unexpected validation message: %q", response.Result.Message)
	}
	if response.Safety.Status != SafetyRejected || response.Safety.Reason != MessageValidationFailed {
		t.Fatalf("unexpected safety decision: %+v", response.Safety)
	}
	if executed {
		t.Fatalf("executor ran for invalid LLM plan")
	}
}

func TestDefaultFormatter(t *testing.T) {
	var results [MaxActions]ActionResult
	formatter := DefaultFormatter{}

	if result := formatter.Format(Plan{}, results, 0, SafetyDecision{}); result.OK {
		t.Fatalf("expected no-result format to fail")
	}

	results[0] = ActionResult{OK: true, Message: MessageOne}
	if result := formatter.Format(Plan{}, results, 1, SafetyDecision{}); result.Message != MessageOne {
		t.Fatalf("expected single result message, got %q", result.Message)
	}

	results[1] = ActionResult{OK: true, Message: MessageTwo}
	if result := formatter.Format(Plan{}, results, 2, SafetyDecision{}); !result.OK || result.Message != MessageCompletedPlan {
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
		{input: "show files", intent: IntentListFiles, action: ActionListFiles},
		{input: "list files", intent: IntentListFiles, action: ActionListFiles},
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

func TestDeterministicPlannerBuildsTargetedFilePlans(t *testing.T) {
	tests := []struct {
		input  string
		intent IntentKind
		action ActionKind
		risk   RiskLevel
		target string
	}{
		{input: "read notes", intent: IntentReadFile, action: ActionReadFile, risk: RiskSafe, target: "notes"},
		{input: "cat notes", intent: IntentReadFile, action: ActionReadFile, risk: RiskSafe, target: "notes"},
		{input: "show notes", intent: IntentReadFile, action: ActionReadFile, risk: RiskSafe, target: "notes"},
		{input: "delete notes", intent: IntentDeleteFile, action: ActionDeleteFile, risk: RiskRisky, target: "notes"},
		{input: "remove notes", intent: IntentDeleteFile, action: ActionDeleteFile, risk: RiskRisky, target: "notes"},
		{input: "stat notes", intent: IntentStatFile, action: ActionStatFile, risk: RiskSafe, target: "notes"},
		{input: "mode deterministic", intent: IntentSetMode, action: ActionSetMode, risk: RiskSafe, target: "deterministic"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := DeterministicPlanner{}.Plan(tt.input, nil)
			if !result.OK {
				t.Fatalf("expected deterministic plan to succeed")
			}
			action := result.Plan.Actions[0]
			if result.Plan.Intent != tt.intent || action.Kind != tt.action || action.Risk != tt.risk {
				t.Fatalf("unexpected plan: intent=%v action=%v risk=%v", result.Plan.Intent, action.Kind, action.Risk)
			}
			if action.TargetLen != len(tt.target) {
				t.Fatalf("target length = %d, expected %d", action.TargetLen, len(tt.target))
			}
			for i := 0; i < action.TargetLen; i++ {
				if action.Target[i] != tt.target[i] {
					t.Fatalf("target byte %d = %q, expected %q", i, action.Target[i], tt.target[i])
				}
			}
		})
	}
}

func TestDeterministicPlannerRecognizesSystemActions(t *testing.T) {
	tests := []struct {
		input  string
		intent IntentKind
		action ActionKind
	}{
		{input: "show history", intent: IntentShowHistory, action: ActionShowHistory},
		{input: "show version", intent: IntentShowVersion, action: ActionShowVersion},
		{input: "show ticks", intent: IntentShowTicks, action: ActionShowTicks},
		{input: "show memory map", intent: IntentShowMemoryMap, action: ActionShowMemoryMap},
		{input: "show memorymap", intent: IntentShowMemoryMap, action: ActionShowMemoryMap},
		{input: "mode", intent: IntentSetMode, action: ActionSetMode},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := DeterministicPlanner{}.Plan(tt.input, nil)
			if !result.OK {
				t.Fatalf("expected deterministic plan to succeed")
			}
			action := result.Plan.Actions[0]
			if result.Plan.Intent != tt.intent || action.Kind != tt.action {
				t.Fatalf("unexpected plan: intent=%v action=%v", result.Plan.Intent, action.Kind)
			}
		})
	}
}

func TestDeterministicPlannerReturnsUnknownForUnsupportedRequests(t *testing.T) {
	result := DeterministicPlanner{}.Plan("make coffee", nil)
	if !result.OK {
		t.Fatalf("expected deterministic planner to return a typed unknown plan")
	}
	if result.Plan.Intent != IntentUnknown {
		t.Fatalf("expected unknown intent, got %v", result.Plan.Intent)
	}
	action := result.Plan.Actions[0]
	if action.Kind != ActionUnknown {
		t.Fatalf("expected unknown action, got %v", action.Kind)
	}
	if action.Risk != RiskSafe {
		t.Fatalf("expected safe risk for unknown action, got %v", action.Risk)
	}
}

func TestDeterministicPlannerLeavesTargetEmptyWhenMissing(t *testing.T) {
	result := DeterministicPlanner{}.Plan("read", nil)
	if !result.OK {
		t.Fatalf("expected deterministic plan to succeed")
	}
	if result.Plan.Actions[0].TargetLen != 0 {
		t.Fatalf("expected empty target, got len=%d", result.Plan.Actions[0].TargetLen)
	}
}

func TestLLMPlannerPropagatesBridgeFailure(t *testing.T) {
	result := LLMPlanner{Bridge: failingBridge{reason: MessageBridgeTimeout}}.Plan("show files", nil)
	if result.OK {
		t.Fatalf("expected bridge failure")
	}
	if result.Reason != MessageBridgeTimeout {
		t.Fatalf("unexpected bridge failure: %q", result.Reason)
	}
}

func TestLLMPlannerDefaultsEmptyBridgeFailureReason(t *testing.T) {
	result := LLMPlanner{Bridge: failingBridge{}}.Plan("show files", nil)
	if result.Reason != MessageLLMBridgeFailed {
		t.Fatalf("unexpected bridge failure: %q", result.Reason)
	}
}

func TestResponseAddTraceStopsAtCapacity(t *testing.T) {
	var response Response
	for i := 0; i < MaxTraceEntries+2; i++ {
		response.AddTrace(TracePlanner, TraceDetailOK)
	}
	if response.TraceCount != MaxTraceEntries {
		t.Fatalf("expected trace count %d, got %d", MaxTraceEntries, response.TraceCount)
	}
}

func TestContextRememberRollsRecentItems(t *testing.T) {
	var context Context
	for i := 0; i < MaxRecentItems+1; i++ {
		context.Remember(IntentShowHelp)
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
	if IntentReadFile.String() != stringReadFile || IntentWriteFile.String() != stringWriteFile || IntentDeleteFile.String() != stringDeleteFile || IntentShowHistory.String() != "show_history" || IntentShowVersion.String() != "show_version" || IntentShowTicks.String() != "show_ticks" || IntentShowMemoryMap.String() != "show_memory_map" || IntentSetMode.String() != "set_mode" || IntentKind(99).String() != stringUnknown {
		t.Fatalf("unexpected intent strings")
	}
	if ActionUnknown.String() != stringUnknown || ActionDeleteFile.String() != stringDeleteFile || ActionShowHistory.String() != "show_history" || ActionShowVersion.String() != "show_version" || ActionShowTicks.String() != "show_ticks" || ActionShowMemoryMap.String() != "show_memory_map" || ActionSetMode.String() != "set_mode" || ActionKind(99).String() != stringUnknown {
		t.Fatalf("unexpected action strings")
	}
	if RiskSafe.String() != "safe" || RiskRisky.String() != "risky" {
		t.Fatalf("unexpected risk strings")
	}
	if SafetyAllowed.String() != "allowed" || SafetyConfirmationRequired.String() != "confirmation_required" || SafetyRejected.String() != "rejected" {
		t.Fatalf("unexpected safety strings")
	}
}

func TestMessageAgentHelpStringMatchesAgentCommands(t *testing.T) {
	want := "Agent commands:\n  agent show files    - Show files managed by the agent\n  agent show history  - Show command history stored by the agent\n  agent show version  - Show OS version through the agent\n  agent show ticks    - Show PIT ticks through the agent\n  agent show memorymap - Show memory map through the agent\n  agent read <name>   - Read a file through the agent\n  agent stat <name>   - Show file metadata through the agent\n  agent delete <name> confirm - Delete a file through the agent\n  agent mode [mode]   - Show or switch agent mode\n  agent help          - Show agent commands"
	if got := MessageAgentHelp.String(); got != want {
		t.Fatalf("MessageAgentHelp.String() = %q, expected %q", got, want)
	}
}

type fakeBridge struct {
	plan Plan
}

func (b fakeBridge) Plan(input string, context *Context) PlanningResult {
	return successfulPlan(b.plan)
}

type failingBridge struct {
	reason MessageKind
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

func planWithTargetAction(kind ActionKind, risk RiskLevel, target string) Plan {
	plan := planWithAction(Action{Kind: kind, Risk: risk})
	targetLen := len(target)
	if targetLen > MaxNameLen {
		targetLen = MaxNameLen
	}
	plan.Actions[0].TargetLen = targetLen
	for i := 0; i < targetLen; i++ {
		plan.Actions[0].Target[i] = target[i]
	}
	return plan
}

func llmPlanWithAction(action Action) Plan {
	plan := planWithAction(action)
	plan.Planner = PlannerModeLLM
	return plan
}

func llmPlanWithTargetAction(kind ActionKind, risk RiskLevel, target string) Plan {
	plan := planWithTargetAction(kind, risk, target)
	plan.Planner = PlannerModeLLM
	return plan
}
