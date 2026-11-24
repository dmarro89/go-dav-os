package agent

func deterministicPlan(input []byte, _ *Context) PlanningResult {
	if containsToken(input, "help") || containsToken(input, "commands") || contains(input, "what can you do") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowHelp, ActionShowHelp, RiskSafe))
	}
	if containsToken(input, "memorymap") || (containsToken(input, "memory") && containsToken(input, "map")) {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowMemoryMap, ActionShowMemoryMap, RiskSafe))
	}
	if containsToken(input, "version") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowVersion, ActionShowVersion, RiskSafe))
	}
	if containsToken(input, "ticks") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowTicks, ActionShowTicks, RiskSafe))
	}
	if containsToken(input, "history") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentShowHistory, ActionShowHistory, RiskSafe))
	}
	if containsToken(input, "mode") {
		return singleTargetPlan(PlannerModeDeterministic, IntentSetMode, ActionSetMode, RiskSafe, input)
	}
	if containsToken(input, "delete") || containsToken(input, "remove") {
		return singleTargetPlan(PlannerModeDeterministic, IntentDeleteFile, ActionDeleteFile, RiskRisky, input)
	}
	if containsToken(input, "read") || containsToken(input, "cat") {
		return singleTargetPlan(PlannerModeDeterministic, IntentReadFile, ActionReadFile, RiskSafe, input)
	}
	if containsToken(input, "stat") || containsToken(input, "status") {
		return singleTargetPlan(PlannerModeDeterministic, IntentStatFile, ActionStatFile, RiskSafe, input)
	}
	if containsToken(input, "show") {
		// "show files" asks for a listing while "show files.txt" names a file.
		if showAsksForList(input) {
			return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentListFiles, ActionListFiles, RiskSafe))
		}
		planning := singleTargetPlan(PlannerModeDeterministic, IntentReadFile, ActionReadFile, RiskSafe, input)
		if !planning.OK || planning.Plan.Actions[0].TargetLen > 0 {
			return planning
		}
	}
	if containsToken(input, "ls") || containsToken(input, "list") || containsToken(input, "files") || containsToken(input, "file") {
		return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentListFiles, ActionListFiles, RiskSafe))
	}
	return successfulPlan(singleActionPlan(PlannerModeDeterministic, IntentUnknown, ActionUnknown, RiskSafe))
}

func singleTargetPlan(mode PlannerMode, intent IntentKind, actionKind ActionKind, risk RiskLevel, input []byte) PlanningResult {
	plan := singleActionPlan(mode, intent, actionKind, risk)
	targetStart, targetEnd := lastToken(input)
	if targetStart < targetEnd && !isActionWord(input, targetStart, targetEnd) {
		targetLen := targetEnd - targetStart
		if targetLen > MaxNameLen {
			return PlanningResult{Reason: MessageActionTargetInvalid}
		}
		plan.Actions[0].TargetLen = targetLen
		for i := 0; i < targetLen; i++ {
			plan.Actions[0].Target[i] = input[targetStart+i]
		}
	}
	return successfulPlan(plan)
}

func successfulPlan(plan Plan) PlanningResult {
	return PlanningResult{OK: true, Plan: plan}
}

func contains(text []byte, token string) bool {
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

// containsToken reports whether token occurs in text as a whole word. Keyword
// matching must not fire inside longer words, otherwise file names such as
// "tools" or "profile" would be misread as list requests.
func containsToken(text []byte, token string) bool {
	for i := 0; i+len(token) <= len(text); i++ {
		if i > 0 && isPlannerWordByte(text[i-1]) {
			continue
		}
		match := true
		for j := 0; j < len(token); j++ {
			if lowerASCII(text[i+j]) != lowerASCII(token[j]) {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		if i+len(token) < len(text) && isPlannerWordByte(text[i+len(token)]) {
			continue
		}
		return true
	}
	return false
}

func isPlannerWordByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// showAsksForList reports whether a whitespace-separated token after "show" is
// a bare list keyword. Word bytes alone cannot tell "files" from "files.txt",
// so the tokens are split on whitespace instead.
func showAsksForList(input []byte) bool {
	seenShow := false
	for i := 0; i < len(input); {
		if isPlannerSpace(input[i]) {
			i++
			continue
		}
		j := i
		for j < len(input) && !isPlannerSpace(input[j]) {
			j++
		}
		if seenShow && isListWord(input, i, j) {
			return true
		}
		if tokenEquals(input, i, j, "show") {
			seenShow = true
		}
		i = j
	}
	return false
}

func isListWord(text []byte, start, end int) bool {
	return tokenEquals(text, start, end, "ls") ||
		tokenEquals(text, start, end, "list") ||
		tokenEquals(text, start, end, "file") ||
		tokenEquals(text, start, end, "files")
}

func lowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

func lastToken(text []byte) (int, int) {
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

func isActionWord(text []byte, start, end int) bool {
	return tokenEquals(text, start, end, "read") ||
		tokenEquals(text, start, end, "cat") ||
		tokenEquals(text, start, end, "delete") ||
		tokenEquals(text, start, end, "remove") ||
		tokenEquals(text, start, end, "show") ||
		tokenEquals(text, start, end, "stat") ||
		tokenEquals(text, start, end, "status") ||
		tokenEquals(text, start, end, "mode")
}

func tokenEquals(text []byte, start, end int, token string) bool {
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
