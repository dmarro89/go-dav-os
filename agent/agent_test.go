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

	var context Context
	response := runtime.Run("delete notes", &context)
	if response.Safety.Status != SafetyConfirmationRequired {
		t.Fatalf("expected confirmation_required, got %v", response.Safety.Status)
	}
	if response.Result.OK {
		t.Fatalf("expected response to stop before execution")
	}
	if executed {
		t.Fatalf("risky action executed without confirmation")
	}
	if !context.ConfirmationPending {
		t.Fatalf("expected risky plan to remain pending")
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
	bridge := fakeBridge{response: bridgeResponse(IntentStatFile, ActionStatFile, RiskSafe, "notes")}
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
	if got := string(plan.Actions[0].Target[:plan.Actions[0].TargetLen]); got != "notes" {
		t.Fatalf("expected converted target, got %q", got)
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

func TestRuntimeExecutesConfirmedDeleteAction(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		DeleteFile: func(action Action, context *Context) ActionResult {
			executed = true
			if action.Risk != RiskRisky {
				t.Fatalf("confirmed action risk = %v, expected risky", action.Risk)
			}
			return ActionResult{OK: true, Message: MessageOK}
		},
	})
	target := [MaxNameLen]byte{'n', 'o', 't', 'e', 's'}
	var context Context

	response := runtime.RunAction(ActionDeleteFile, IntentDeleteFile, RiskRisky, &target, 5, &context)
	if response.Safety.Status != SafetyConfirmationRequired || executed {
		t.Fatalf("expected delete to wait for confirmation")
	}
	if context.CurrentTask != ActionDeleteFile || context.LastResultSummary != MessageConfirmationRequired {
		t.Fatalf("pending context was not recorded: %+v", context)
	}

	response = runtime.ConfirmAction(true, &context)
	if !response.Result.OK || !executed {
		t.Fatalf("expected confirmed delete action to execute, got %+v", response.Result)
	}
	if context.CurrentTask != ActionNone || context.LastResultSummary != MessageOK {
		t.Fatalf("confirmed context was not completed: %+v", context)
	}
}

func TestRunActionMessageRejectsTooLongTarget(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		DeleteFile: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageOK}
		},
	})
	target := [MaxNameLen]byte{}
	var context Context

	message := runtime.RunActionMessage(ActionDeleteFile, IntentDeleteFile, RiskRisky, &target, MaxNameLen+1, &context)

	if message != MessageActionTargetInvalid {
		t.Fatalf("expected invalid target message, got %v", message)
	}
	if executed || context.ConfirmationPending {
		t.Fatalf("invalid action must not be queued or executed")
	}
}

func TestRuntimeCancelsPendingAction(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		DeleteFile: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageOK}
		},
	})
	var context Context

	runtime.Run("delete notes", &context)
	response := runtime.ConfirmAction(false, &context)

	if response.Result.Message != MessageActionCancelled || executed || context.ConfirmationPending {
		t.Fatalf("expected pending action to be cancelled")
	}
}

func TestLLMPlanUsesSameConfirmationGate(t *testing.T) {
	executed := false
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		DeleteFile: func(action Action, context *Context) ActionResult {
			executed = true
			return ActionResult{OK: true, Message: MessageOK}
		},
	})
	var context Context
	plan := llmPlanWithTargetAction(ActionDeleteFile, RiskRisky, "notes")
	plan.Intent = IntentDeleteFile

	response := runtime.runPlan(plan, &context)
	if response.Safety.Status != SafetyConfirmationRequired || executed {
		t.Fatalf("expected LLM plan to wait for confirmation")
	}
	if response = runtime.ConfirmAction(true, &context); !response.Result.OK || !executed {
		t.Fatalf("expected confirmed LLM plan to execute")
	}
}

func TestRuntimeStopsOnExecutorFailure(t *testing.T) {
	var context Context
	response := NewDeterministicAgent(
		AllowedActionExecutor{
			ReadFile: func(action Action, context *Context) ActionResult {
				return ActionResult{OK: false, Message: MessageReadFailed}
			},
		},
	).Run("read notes", &context)

	if response.Result.OK {
		t.Fatalf("expected executor failure")
	}
	if response.Result.Message != MessageReadFailed {
		t.Fatalf("unexpected executor failure: %q", response.Result.Message)
	}
	if context.LastAction != ActionReadFile || context.LastResultSummary != MessageReadFailed || context.RequestCount != 1 {
		t.Fatalf("executor failure was not recorded in context: %+v", context)
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

func TestRuntimePlannerModeSelection(t *testing.T) {
	listCalls := 0
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		ListFiles: func(action Action, context *Context) ActionResult {
			listCalls++
			return ActionResult{OK: true, Message: MessageFilesListed}
		},
	})

	if current := runtime.CurrentPlanner(); !current.OK || current.Message != MessageDeterministicMode {
		t.Fatalf("unexpected initial planner: %+v", current)
	}
	if result := runtime.SetPlannerMode(PlannerModeLLM); result.OK || result.Message != MessageLLMModeNotConfigured {
		t.Fatalf("unexpected unavailable LLM result: %+v", result)
	}
	if runtime.PlannerMode() != PlannerModeDeterministic {
		t.Fatalf("failed switch changed planner mode to %v", runtime.PlannerMode())
	}
	if response := runtime.Run("show files", nil); !response.Result.OK || listCalls != 1 {
		t.Fatalf("deterministic planner did not remain usable: %+v", response.Result)
	}

	runtime.ConfigureLLMPlanner(LLMPlanner{
		Bridge: fakeBridge{response: bridgeResponse(IntentListFiles, ActionListFiles, RiskSafe, "")},
	})
	if result := runtime.SetPlannerMode(PlannerModeLLM); !result.OK || result.Message != MessagePlannerSwitchedLLM {
		t.Fatalf("unexpected LLM switch result: %+v", result)
	}
	var context Context
	response := runtime.Run("show files", &context)
	if !response.Result.OK || listCalls != 2 || context.PlannerMode != PlannerModeLLM {
		t.Fatalf("LLM planner did not execute: response=%+v context=%+v", response.Result, context)
	}
	if response.TraceCount == 0 || response.Trace[0].Detail != TraceDetailLLM {
		t.Fatalf("expected LLM planner trace, got %+v", response.Trace[0])
	}

	if result := runtime.SetPlannerMode(PlannerModeDeterministic); !result.OK || result.Message != MessagePlannerSwitchedDeterministic {
		t.Fatalf("unexpected deterministic switch result: %+v", result)
	}
	if current := runtime.CurrentPlanner(); current.Message != MessageDeterministicMode {
		t.Fatalf("unexpected final planner: %+v", current)
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

func TestLLMPlannerRejectsInvalidBridgeResponses(t *testing.T) {
	longTarget := bridgeResponse(IntentReadFile, ActionReadFile, RiskSafe, "notes")
	longTarget.TargetLen = MaxNameLen + 1
	tests := []struct {
		name     string
		response BridgeResponse
		reason   MessageKind
	}{
		{
			name:     "action outside allowlist",
			response: bridgeResponse(IntentWriteFile, ActionWriteFile, RiskRisky, "notes"),
			reason:   MessagePlanContainsUnsupportedAction,
		},
		{
			name:     "missing target",
			response: bridgeResponse(IntentReadFile, ActionReadFile, RiskSafe, ""),
			reason:   MessageActionTargetInvalid,
		},
		{
			name:     "incorrect risk",
			response: bridgeResponse(IntentDeleteFile, ActionDeleteFile, RiskSafe, "notes"),
			reason:   MessageActionRiskInvalid,
		},
		{
			name:     "mismatched intent",
			response: bridgeResponse(IntentDeleteFile, ActionReadFile, RiskSafe, "notes"),
			reason:   MessagePlannerFailed,
		},
		{
			name:     "target beyond limit",
			response: longTarget,
			reason:   MessageActionTargetInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LLMPlanner{Bridge: fakeBridge{response: tt.response}}.Plan("request", nil)
			if result.OK || result.Reason != tt.reason {
				t.Fatalf("unexpected planning result: %+v", result)
			}
		})
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

func TestRuntimeUpdatesSessionContext(t *testing.T) {
	runtime := NewDeterministicAgent(AllowedActionExecutor{
		ListFiles: func(action Action, context *Context) ActionResult {
			if context.CurrentTask != ActionListFiles {
				t.Fatalf("executor current task = %v, expected list files", context.CurrentTask)
			}
			return ActionResult{OK: true, Message: MessageFilesListed}
		},
	})
	var input [MaxContextInput]byte
	copy(input[:], "agent show files")
	var context Context

	response := runtime.RunActionRequest(
		ActionListFiles,
		IntentListFiles,
		RiskSafe,
		nil,
		0,
		&input,
		len("agent show files"),
		&context,
	)

	if !response.Result.OK {
		t.Fatalf("expected successful response, got %+v", response)
	}
	if context.CurrentTask != ActionNone ||
		context.LastIntent != IntentListFiles ||
		context.LastAction != ActionListFiles ||
		context.LastResultSummary != MessageFilesListed ||
		context.RequestCount != 1 ||
		context.PlannerMode != PlannerModeDeterministic ||
		context.RecentCount != 1 {
		t.Fatalf("unexpected context: %+v", context)
	}
	if got := string(context.LastInput[:context.LastInputLen]); got != "agent show files" {
		t.Fatalf("last input = %q, expected %q", got, "agent show files")
	}
}

func TestLLMPlannerForwardsSessionContext(t *testing.T) {
	bridge := &contextBridge{
		response: bridgeResponse(IntentListFiles, ActionListFiles, RiskSafe, ""),
	}
	context := Context{
		LastIntent:        IntentReadFile,
		LastAction:        ActionReadFile,
		LastResultSummary: MessageFileRead,
		RequestCount:      3,
		PlannerMode:       PlannerModeDeterministic,
	}

	result := (LLMPlanner{Bridge: bridge}).Plan("show files", &context)

	if !result.OK {
		t.Fatalf("expected successful plan, got %+v", result)
	}
	if bridge.request.Context != &context {
		t.Fatalf("bridge did not receive the session context")
	}
	if bridge.request.Input != "show files" {
		t.Fatalf("bridge input = %q, expected %q", bridge.request.Input, "show files")
	}
	if bridge.request.Context.RequestCount != 3 || bridge.request.Context.LastAction != ActionReadFile {
		t.Fatalf("bridge received unexpected context: %+v", bridge.request.Context)
	}
	if bridge.request.AllowedActionCount != MaxBridgeAllowedActions {
		t.Fatalf("bridge allowed action count = %d", bridge.request.AllowedActionCount)
	}
	for i := 0; i < bridge.request.AllowedActionCount; i++ {
		if bridge.request.AllowedActions[i] == ActionWriteFile || bridge.request.AllowedActions[i] == ActionSetMode {
			t.Fatalf("bridge received unsupported action %v", bridge.request.AllowedActions[i])
		}
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
	want := "Agent commands:\n  agent show files    - Show files managed by the agent\n  agent show history  - Show command history stored by the agent\n  agent show version  - Show OS version through the agent\n  agent show ticks    - Show PIT ticks through the agent\n  agent show memorymap - Show memory map through the agent\n  agent read <name>   - Read a file through the agent\n  agent stat <name>   - Show file metadata through the agent\n  agent delete <name> - Delete a file after confirmation\n  agent mode [mode]   - Show or switch agent mode\n  agent transport ping - Probe the guest-host transport\n  agent context       - Show current agent context\n  agent help          - Show agent commands"
	if got := MessageAgentHelp.String(); got != want {
		t.Fatalf("MessageAgentHelp.String() = %q, expected %q", got, want)
	}
}

type fakeBridge struct {
	response BridgeResponse
}

func (b fakeBridge) Plan(request BridgeRequest) BridgeResult {
	return BridgeResult{OK: true, Response: b.response}
}

type contextBridge struct {
	response BridgeResponse
	request  BridgeRequest
}

func (b *contextBridge) Plan(request BridgeRequest) BridgeResult {
	b.request = request
	return BridgeResult{OK: true, Response: b.response}
}

type failingBridge struct {
	reason MessageKind
}

func (b failingBridge) Plan(request BridgeRequest) BridgeResult {
	return BridgeResult{OK: false, Reason: b.reason}
}

func bridgeResponse(intent IntentKind, action ActionKind, risk RiskLevel, target string) BridgeResponse {
	var response BridgeResponse
	response.Intent = intent
	response.Action = action
	response.Risk = risk
	response.TargetLen = len(target)
	if response.TargetLen > MaxNameLen {
		response.TargetLen = MaxNameLen
	}
	for i := 0; i < response.TargetLen; i++ {
		response.Target[i] = target[i]
	}
	return response
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
