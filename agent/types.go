package agent

const (
	MaxActions      = 4
	MaxTraceEntries = 8
	MaxNameLen      = 16
	MaxDataLen      = 128
	MaxRecentItems  = 4
)

type PlannerMode uint8

const (
	PlannerModeDeterministic PlannerMode = iota
	PlannerModeLLM
)

type IntentKind uint8

const (
	IntentUnknown IntentKind = iota
	IntentListFiles
	IntentReadFile
	IntentWriteFile
	IntentDeleteFile
	IntentStatFile
	IntentShowHelp
	IntentShowHistory
	IntentSetMode
)

type ActionKind uint8

const (
	ActionNone ActionKind = iota
	ActionListFiles
	ActionReadFile
	ActionWriteFile
	ActionDeleteFile
	ActionStatFile
	ActionShowHelp
	ActionShowHistory
	ActionSetMode
)

func (k ActionKind) Valid() bool {
	switch k {
	case ActionListFiles, ActionReadFile, ActionWriteFile, ActionDeleteFile, ActionStatFile, ActionShowHelp, ActionShowHistory, ActionSetMode:
		return true
	default:
		return false
	}
}

type RiskLevel uint8

const (
	RiskSafe RiskLevel = iota
	RiskRisky
)

type SafetyStatus uint8

const (
	SafetyRejected SafetyStatus = iota
	SafetyAllowed
	SafetyConfirmationRequired
)

type MessageKind uint8

const (
	MessageNone MessageKind = iota
	MessagePlannerFailed
	MessagePlannerMissing
	MessageValidationFailed
	MessagePlanHasNoActions
	MessagePlanHasTooManyActions
	MessagePlanContainsUnsupportedAction
	MessageActionRiskInvalid
	MessageActionTargetInvalid
	MessageActionDataInvalid
	MessageConfirmationRequired
	MessageExecutorNotConfigured
	MessageExecutorMissing
	MessageUnsupportedAction
	MessageActionUnavailable
	MessageNoResult
	MessageCompletedPlan
	MessageOK
	MessageFilesListed
	MessageNoFiles
	MessageFileRead
	MessageFileStat
	MessageMissingFile
	MessageFileNotFound
	MessageAgentHelp
	MessageHistoryListed
	MessageDeterministicMode
	MessageLLMModeNotConfigured
	MessageUnsupportedMode
	MessageReadFailed
	MessageHelp
	MessageOne
	MessageTwo
	MessageLLMBridgeNotConfigured
	MessageLLMBridgeFailed
	MessageBridgeTimeout
)

type TraceKind uint8

const (
	TraceNone TraceKind = iota
	TracePlanner
	TraceIntent
	TraceValidation
	TraceSafety
	TraceExecutor
	TraceFormatter
)

type TraceDetail uint8

const (
	TraceDetailNone TraceDetail = iota
	TraceDetailMissing
	TraceDetailFailed
	TraceDetailOK
	TraceDetailAllowed
	TraceDetailRejected
	TraceDetailConfirmationRequired
	TraceDetailSuccess
	TraceDetailStructured
	TraceDetailDeterministic
	TraceDetailLLM
	TraceDetailListFiles
	TraceDetailReadFile
	TraceDetailWriteFile
	TraceDetailDeleteFile
	TraceDetailStatFile
	TraceDetailShowHelp
	TraceDetailShowHistory
	TraceDetailSetMode
)

// Action is the only executable unit the agent runtime understands.
type Action struct {
	Kind      ActionKind
	Risk      RiskLevel
	Target    [MaxNameLen]byte
	TargetLen int
	Data      [MaxDataLen]byte
	DataLen   int
}

type Plan struct {
	Planner     PlannerMode
	Intent      IntentKind
	Confidence  uint8
	Actions     [MaxActions]Action
	ActionCount int
}

type PlanningResult struct {
	OK     bool
	Plan   Plan
	Reason MessageKind
}

type ValidationResult struct {
	OK     bool
	Reason MessageKind
}

type SafetyDecision struct {
	Status SafetyStatus
	Reason MessageKind
}

type ActionResult struct {
	OK      bool
	Message MessageKind
}

type TraceEntry struct {
	Stage  TraceKind
	Detail TraceDetail
}

type Response struct {
	Trace      [MaxTraceEntries]TraceEntry
	TraceCount int
	Result     ActionResult
	Safety     SafetyDecision
}

func (r *Response) AddTrace(stage TraceKind, detail TraceDetail) {
	if r.TraceCount >= MaxTraceEntries {
		return
	}
	r.Trace[r.TraceCount].Stage = stage
	r.Trace[r.TraceCount].Detail = detail
	r.TraceCount++
}

type Context struct {
	LastIntent  IntentKind
	RecentCount int
}

func (c *Context) Remember(intent IntentKind) {
	if c == nil {
		return
	}
	c.LastIntent = intent
	if c.RecentCount < MaxRecentItems {
		c.RecentCount++
	}
}
