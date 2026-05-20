//go:build !gccgo

package agent

func deterministicPlan(input string, _ *Context) PlanningResult {
	if contains(input, "help") || contains(input, "commands") || contains(input, "what can you do") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowHelp, ActionShowHelp, RiskSafe))
	}
	if contains(input, "memorymap") || (contains(input, "memory") && contains(input, "map")) {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowMemoryMap, ActionShowMemoryMap, RiskSafe))
	}
	if contains(input, "version") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowVersion, ActionShowVersion, RiskSafe))
	}
	if contains(input, "ticks") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowTicks, ActionShowTicks, RiskSafe))
	}
	if contains(input, "history") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowHistory, ActionShowHistory, RiskSafe))
	}
	if contains(input, "mode") {
		return successfulPlan(singleTargetPlan(PlannerModeDeterministic, IntentSetMode, ActionSetMode, RiskSafe, input))
	}
	if contains(input, "delete") || contains(input, "remove") {
		return successfulPlan(singleTargetPlan(PlannerModeDeterministic, IntentDeleteFile, ActionDeleteFile, RiskRisky, input))
	}
	if contains(input, "read") || contains(input, "cat") {
		return successfulPlan(singleTargetPlan(PlannerModeDeterministic, IntentReadFile, ActionReadFile, RiskSafe, input))
	}
	if contains(input, "stat") || contains(input, "status") {
		return successfulPlan(singleTargetPlan(PlannerModeDeterministic, IntentStatFile, ActionStatFile, RiskSafe, input))
	}
	if contains(input, "ls") || contains(input, "list") || contains(input, "files") || contains(input, "file") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentListFiles, ActionListFiles, RiskSafe))
	}
	if contains(input, "show") {
		plan := singleTargetPlan(PlannerModeDeterministic, IntentReadFile, ActionReadFile, RiskSafe, input)
		if plan.Actions[0].TargetLen > 0 {
			return successfulPlan(plan)
		}
	}
	return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentUnknown, ActionUnknown, RiskSafe))
}

func singleTargetPlan(mode PlannerMode, intent IntentKind, actionKind ActionKind, risk RiskLevel, input string) Plan {
	plan := singleActionPlan(mode, intent, actionKind, risk)
	targetStart, targetEnd := lastToken(input)
	if targetStart < targetEnd && !isActionWord(input, targetStart, targetEnd) {
		targetLen := targetEnd - targetStart
		if targetLen > MaxNameLen {
			targetLen = MaxNameLen
		}
		plan.Actions[0].TargetLen = targetLen
		for i := 0; i < targetLen; i++ {
			plan.Actions[0].Target[i] = input[targetStart+i]
		}
	}
	return plan
}

func successfulPlan(plan Plan) PlanningResult {
	return PlanningResult{OK: true, Plan: plan}
}

func contains(text, token string) bool {
	if len(token) == 0 {
		return true
	}
	if len(token) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(token); i++ {
		match := true
		for j := 0; j < len(token); j++ {
			if lowerASCII(text[i+j]) != lowerASCII(token[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func lowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

func lastToken(text string) (int, int) {
	end := len(text)
	for end > 0 && isPlannerSpace(text[end-1]) {
		end--
	}
	start := end
	for start > 0 && !isPlannerSpace(text[start-1]) {
		start--
	}
	return start, end
}

func isPlannerSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isActionWord(text string, start, end int) bool {
	return tokenEquals(text, start, end, "read") ||
		tokenEquals(text, start, end, "cat") ||
		tokenEquals(text, start, end, "delete") ||
		tokenEquals(text, start, end, "remove") ||
		tokenEquals(text, start, end, "show") ||
		tokenEquals(text, start, end, "stat") ||
		tokenEquals(text, start, end, "status") ||
		tokenEquals(text, start, end, "mode")
}

func tokenEquals(text string, start, end int, token string) bool {
	if end-start != len(token) {
		return false
	}
	for i := 0; i < len(token); i++ {
		if lowerASCII(text[start+i]) != token[i] {
			return false
		}
	}
	return true
}
