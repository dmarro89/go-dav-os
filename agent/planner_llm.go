//go:build !gccgo

package agent

const MaxBridgeAllowedActions = 8

type BridgeRequest struct {
	Input              string
	Context            *Context
	AllowedActions     [MaxBridgeAllowedActions]ActionKind
	AllowedActionCount int
}

type BridgeResponse struct {
	Intent    IntentKind
	Action    ActionKind
	Risk      RiskLevel
	Target    [MaxNameLen]byte
	TargetLen int
}

type BridgeResult struct {
	OK       bool
	Response BridgeResponse
	Reason   MessageKind
}

type BridgeClient interface {
	Plan(request BridgeRequest) BridgeResult
}

type LLMPlanner struct {
	Bridge BridgeClient
}

var _ Planner = LLMPlanner{}

const errLLMBridgeNotConfigured = "agent: llm bridge not configured"

func (p LLMPlanner) Plan(input string, context *Context) PlanningResult {
	if p.Bridge == nil {
		return PlanningResult{OK: false, Reason: MessageLLMBridgeNotConfigured}
	}

	request := newBridgeRequest(input, context)
	result := p.Bridge.Plan(request)
	if !result.OK {
		if result.Reason == MessageNone {
			result.Reason = MessageLLMBridgeFailed
		}
		return PlanningResult{OK: false, Reason: result.Reason}
	}
	if !request.allows(result.Response.Action) {
		return PlanningResult{OK: false, Reason: MessagePlanContainsUnsupportedAction}
	}
	if bridgeActionIntent(result.Response.Action) != result.Response.Intent {
		return PlanningResult{OK: false, Reason: MessagePlannerFailed}
	}

	plan := singleActionPlan(PlannerModeLLM, result.Response.Intent, result.Response.Action, result.Response.Risk)
	plan.Actions[0].Target = result.Response.Target
	plan.Actions[0].TargetLen = result.Response.TargetLen
	validation := validatePlan(plan)
	if !validation.OK {
		return PlanningResult{OK: false, Reason: validation.Reason}
	}
	return successfulPlan(plan)
}

func newBridgeRequest(input string, context *Context) BridgeRequest {
	return BridgeRequest{
		Input:   input,
		Context: context,
		AllowedActions: [MaxBridgeAllowedActions]ActionKind{
			ActionListFiles,
			ActionReadFile,
			ActionStatFile,
			ActionDeleteFile,
			ActionShowHistory,
			ActionShowVersion,
			ActionShowTicks,
			ActionShowMemoryMap,
		},
		AllowedActionCount: MaxBridgeAllowedActions,
	}
}

func (r BridgeRequest) allows(kind ActionKind) bool {
	for i := 0; i < r.AllowedActionCount; i++ {
		if r.AllowedActions[i] == kind {
			return true
		}
	}
	return false
}

func bridgeActionIntent(kind ActionKind) IntentKind {
	switch kind {
	case ActionListFiles:
		return IntentListFiles
	case ActionReadFile:
		return IntentReadFile
	case ActionStatFile:
		return IntentStatFile
	case ActionDeleteFile:
		return IntentDeleteFile
	case ActionShowHistory:
		return IntentShowHistory
	case ActionShowVersion:
		return IntentShowVersion
	case ActionShowTicks:
		return IntentShowTicks
	case ActionShowMemoryMap:
		return IntentShowMemoryMap
	default:
		return IntentUnknown
	}
}
