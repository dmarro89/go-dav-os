//go:build !gccgo

package agent

const (
	stringUnknown              = "unknown"
	stringDeterministic        = "deterministic"
	stringLLM                  = "llm"
	stringListFiles            = "list_files"
	stringReadFile             = "read_file"
	stringWriteFile            = "write_file"
	stringDeleteFile           = "delete_file"
	stringStatFile             = "stat_file"
	stringShowHelp             = "show_help"
	stringShowHistory          = "show_history"
	stringSetMode              = "set_mode"
	stringAllowed              = "allowed"
	stringConfirmationRequired = "confirmation_required"
	stringRejected             = "rejected"
	stringNone                 = "none"
)

func (m PlannerMode) String() string {
	switch m {
	case PlannerModeDeterministic:
		return stringDeterministic
	case PlannerModeLLM:
		return stringLLM
	default:
		return stringUnknown
	}
}

func (i IntentKind) String() string {
	switch i {
	case IntentListFiles:
		return stringListFiles
	case IntentReadFile:
		return stringReadFile
	case IntentWriteFile:
		return stringWriteFile
	case IntentDeleteFile:
		return stringDeleteFile
	case IntentStatFile:
		return stringStatFile
	case IntentShowHelp:
		return stringShowHelp
	case IntentShowHistory:
		return stringShowHistory
	case IntentSetMode:
		return stringSetMode
	default:
		return stringUnknown
	}
}

func (k ActionKind) String() string {
	switch k {
	case ActionListFiles:
		return stringListFiles
	case ActionReadFile:
		return stringReadFile
	case ActionWriteFile:
		return stringWriteFile
	case ActionDeleteFile:
		return stringDeleteFile
	case ActionStatFile:
		return stringStatFile
	case ActionShowHelp:
		return stringShowHelp
	case ActionShowHistory:
		return stringShowHistory
	case ActionSetMode:
		return stringSetMode
	default:
		return stringNone
	}
}

func (r RiskLevel) String() string {
	switch r {
	case RiskRisky:
		return "risky"
	default:
		return "safe"
	}
}

func (s SafetyStatus) String() string {
	switch s {
	case SafetyAllowed:
		return stringAllowed
	case SafetyConfirmationRequired:
		return stringConfirmationRequired
	default:
		return stringRejected
	}
}

func (m MessageKind) String() string {
	switch m {
	case MessagePlannerFailed:
		return "agent: planner failed"
	case MessagePlannerMissing:
		return "planner_missing"
	case MessageValidationFailed:
		return "validation_failed"
	case MessagePlanHasNoActions:
		return "agent: plan has no actions"
	case MessagePlanHasTooManyActions:
		return "agent: plan has too many actions"
	case MessagePlanContainsUnsupportedAction:
		return "agent: plan contains unsupported action"
	case MessageActionRiskInvalid:
		return "agent: action risk is invalid"
	case MessageActionTargetInvalid:
		return "agent: action target is invalid"
	case MessageActionDataInvalid:
		return "agent: action data is invalid"
	case MessageConfirmationRequired:
		return "agent: confirmation required"
	case MessageExecutorNotConfigured:
		return "agent: executor not configured"
	case MessageExecutorMissing:
		return "missing"
	case MessageUnsupportedAction:
		return "agent: unsupported action"
	case MessageActionUnavailable:
		return "agent: action unavailable"
	case MessageNoResult:
		return "agent: no result"
	case MessageCompletedPlan:
		return "agent: completed plan"
	case MessageOK:
		return "ok"
	case MessageFilesListed:
		return "agent: files listed"
	case MessageNoFiles:
		return "agent: no files"
	case MessageFileRead:
		return "agent: file read"
	case MessageFileStat:
		return "agent: file stat"
	case MessageMissingFile:
		return "agent: missing file"
	case MessageFileNotFound:
		return "agent: file not found"
	case MessageAgentHelp:
		return "Agent commands:\n  agent show files    - Show files managed by the agent\n  agent show history  - Show command history stored by the agent\n  agent read <name>   - Read a file through the agent\n  agent stat <name>   - Show file metadata through the agent\n  agent delete <name> - Request deletion through the agent\n  agent mode <mode>   - Switch agent mode (not implemented yet)"
	case MessageHistoryListed:
		return "agent: history listed"
	case MessageDeterministicMode:
		return "agent: deterministic mode"
	case MessageLLMModeNotConfigured:
		return "agent: llm mode not configured"
	case MessageUnsupportedMode:
		return "agent: unsupported mode"
	case MessageReadFailed:
		return "read failed"
	case MessageHelp:
		return "help"
	case MessageOne:
		return "one"
	case MessageTwo:
		return "two"
	case MessageLLMBridgeNotConfigured:
		return errLLMBridgeNotConfigured
	case MessageLLMBridgeFailed:
		return "agent: llm bridge failed"
	case MessageBridgeTimeout:
		return "bridge timeout"
	default:
		return ""
	}
}

func (t TraceKind) String() string {
	switch t {
	case TracePlanner:
		return "Planner"
	case TraceIntent:
		return "Intent"
	case TraceValidation:
		return "Validation"
	case TraceSafety:
		return "Safety"
	case TraceExecutor:
		return "Executor"
	case TraceFormatter:
		return "Formatter"
	default:
		return ""
	}
}

func (d TraceDetail) String() string {
	switch d {
	case TraceDetailMissing:
		return "missing"
	case TraceDetailFailed:
		return "failed"
	case TraceDetailOK:
		return "ok"
	case TraceDetailAllowed:
		return stringAllowed
	case TraceDetailRejected:
		return stringRejected
	case TraceDetailConfirmationRequired:
		return stringConfirmationRequired
	case TraceDetailSuccess:
		return "success"
	case TraceDetailStructured:
		return "structured"
	case TraceDetailDeterministic:
		return stringDeterministic
	case TraceDetailLLM:
		return stringLLM
	case TraceDetailListFiles:
		return stringListFiles
	case TraceDetailReadFile:
		return stringReadFile
	case TraceDetailWriteFile:
		return stringWriteFile
	case TraceDetailDeleteFile:
		return stringDeleteFile
	case TraceDetailStatFile:
		return stringStatFile
	case TraceDetailShowHelp:
		return stringShowHelp
	case TraceDetailShowHistory:
		return stringShowHistory
	case TraceDetailSetMode:
		return stringSetMode
	default:
		return ""
	}
}
