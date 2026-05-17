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

	stringUnknown    = "unknown"
	stringReadFile   = "read_file"
	stringWriteFile  = "write_file"
	stringDeleteFile = "delete_file"
	stringNone       = "none"
)

func (m PlannerMode) String() string {
	switch m {
	case PlannerModeDeterministic:
		return "deterministic"
	case PlannerModeLLM:
		return "llm"
	default:
		return stringUnknown
	}
}

type IntentKind uint8

const (
	IntentUnknown IntentKind = iota
	IntentListFiles
	IntentReadFile
	IntentWriteFile
	IntentDeleteFile
	IntentStatFile
	IntentShowHelp
)

func (i IntentKind) String() string {
	switch i {
	case IntentListFiles:
		return "list_files"
	case IntentReadFile:
		return stringReadFile
	case IntentWriteFile:
		return stringWriteFile
	case IntentDeleteFile:
		return stringDeleteFile
	case IntentStatFile:
		return "stat_file"
	case IntentShowHelp:
		return "show_help"
	default:
		return stringUnknown
	}
}

type ActionKind uint8

const (
	ActionNone ActionKind = iota
	ActionListFiles
	ActionReadFile
	ActionWriteFile
	ActionDeleteFile
	ActionStatFile
	ActionShowHelp
)

func (k ActionKind) String() string {
	switch k {
	case ActionListFiles:
		return "list_files"
	case ActionReadFile:
		return stringReadFile
	case ActionWriteFile:
		return stringWriteFile
	case ActionDeleteFile:
		return stringDeleteFile
	case ActionStatFile:
		return "stat_file"
	case ActionShowHelp:
		return "show_help"
	default:
		return stringNone
	}
}

type RiskLevel uint8

const (
	RiskSafe RiskLevel = iota
	RiskRisky
)

func (r RiskLevel) String() string {
	switch r {
	case RiskRisky:
		return "risky"
	default:
		return "safe"
	}
}

type SafetyStatus uint8

const (
	SafetyRejected SafetyStatus = iota
	SafetyAllowed
	SafetyConfirmationRequired
)

func (s SafetyStatus) String() string {
	switch s {
	case SafetyAllowed:
		return "allowed"
	case SafetyConfirmationRequired:
		return "confirmation_required"
	default:
		return "rejected"
	}
}

// Action is the only executable unit the agent runtime understands
// There is deliberately no raw command field here: shell integration must map
// these typed action kinds onto explicit internal APIs in a later layer
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
	Reason string
}

type ValidationResult struct {
	OK     bool
	Reason string
}

type SafetyDecision struct {
	Status SafetyStatus
	Reason string
}

type ActionResult struct {
	OK      bool
	Message string
}

type TraceEntry struct {
	Stage  string
	Detail string
}

type Response struct {
	Trace      [MaxTraceEntries]TraceEntry
	TraceCount int
	Result     ActionResult
	Safety     SafetyDecision
}

func (r *Response) AddTrace(stage, detail string) {
	if r.TraceCount >= MaxTraceEntries {
		return
	}
	r.Trace[r.TraceCount] = TraceEntry{Stage: stage, Detail: detail}
	r.TraceCount++
}

type Context struct {
	LastIntent    IntentKind
	RecentInputs  [MaxRecentItems]string
	RecentResults [MaxRecentItems]string
	RecentCount   int
}

func (c *Context) Remember(input, result string, intent IntentKind) {
	if c == nil {
		return
	}
	c.LastIntent = intent
	if c.RecentCount < MaxRecentItems {
		c.RecentInputs[c.RecentCount] = input
		c.RecentResults[c.RecentCount] = result
		c.RecentCount++
		return
	}
	for i := 1; i < MaxRecentItems; i++ {
		c.RecentInputs[i-1] = c.RecentInputs[i]
		c.RecentResults[i-1] = c.RecentResults[i]
	}
	c.RecentInputs[MaxRecentItems-1] = input
	c.RecentResults[MaxRecentItems-1] = result
}
