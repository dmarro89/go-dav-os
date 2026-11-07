package agent

const MaxBridgeAllowedActions = 8

type BridgeRequest struct {
	Input              [MaxContextInput]byte
	InputLen           int
	Context            Context
	HasContext         bool
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

func NewBridgeRequest(input *[MaxContextInput]byte, inputLen int, context *Context) BridgeRequest {
	var request BridgeRequest
	if inputLen < 0 {
		inputLen = 0
	}
	if inputLen > MaxContextInput {
		inputLen = MaxContextInput
	}
	if input == nil {
		inputLen = 0
	}
	for i := 0; i < inputLen; i++ {
		request.Input[i] = input[i]
	}
	request.InputLen = inputLen
	if context != nil {
		request.Context = *context
		request.HasContext = true
	}
	request.AllowedActions = [MaxBridgeAllowedActions]ActionKind{
		ActionListFiles,
		ActionReadFile,
		ActionStatFile,
		ActionDeleteFile,
		ActionShowHistory,
		ActionShowVersion,
		ActionShowTicks,
		ActionShowMemoryMap,
	}
	request.AllowedActionCount = MaxBridgeAllowedActions
	return request
}

func bridgePlanningResult(request BridgeRequest, result BridgeResult) PlanningResult {
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
	if !actionRequiresTarget(result.Response.Action) && result.Response.TargetLen != 0 {
		return PlanningResult{OK: false, Reason: MessageActionTargetInvalid}
	}

	plan := singleActionPlan(PlannerModeLLM, result.Response.Intent, result.Response.Action, result.Response.Risk)
	plan.Actions[0].Target = result.Response.Target
	plan.Actions[0].TargetLen = result.Response.TargetLen
	validation := validatePlan(plan)
	if !validation.OK {
		return PlanningResult{OK: false, Reason: validation.Reason}
	}
	return PlanningResult{OK: true, Plan: plan, Reason: MessageOK}
}

func (r BridgeRequest) allows(kind ActionKind) bool {
	if r.AllowedActionCount < 0 || r.AllowedActionCount > MaxBridgeAllowedActions {
		return false
	}
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
