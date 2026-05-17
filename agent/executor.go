package agent

type ActionHandler func(action Action, context *Context) ActionResult

// AllowedActionExecutor is the constrained action surface for the agent
// Shell or filesystem integration can wire only the handlers it wants to expose;
// unsupported action kinds fail closed instead of falling through to a shell
type AllowedActionExecutor struct {
	ListFiles  ActionHandler
	ReadFile   ActionHandler
	WriteFile  ActionHandler
	DeleteFile ActionHandler
	StatFile   ActionHandler
	ShowHelp   ActionHandler
}

func (e AllowedActionExecutor) Execute(action Action, context *Context) ActionResult {
	switch action.Kind {
	case ActionListFiles:
		return callHandler(e.ListFiles, action, context)
	case ActionReadFile:
		return callHandler(e.ReadFile, action, context)
	case ActionWriteFile:
		return callHandler(e.WriteFile, action, context)
	case ActionDeleteFile:
		return callHandler(e.DeleteFile, action, context)
	case ActionStatFile:
		return callHandler(e.StatFile, action, context)
	case ActionShowHelp:
		return callHandler(e.ShowHelp, action, context)
	default:
		return ActionResult{OK: false, Message: "agent: unsupported action"}
	}
}

func callHandler(handler ActionHandler, action Action, context *Context) ActionResult {
	if handler == nil {
		return ActionResult{OK: false, Message: "agent: action unavailable"}
	}
	return handler(action, context)
}
