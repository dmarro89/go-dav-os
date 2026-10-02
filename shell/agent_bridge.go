package shell

import (
	"github.com/dmarro89/go-dav-os/agent"
	"github.com/dmarro89/go-dav-os/serial"
)

func serialBridgeAvailable() bool {
	const healthRequest = `{"health":"ping"}`
	requestLen := copyLiteral(&transportRequest, healthRequest)
	responseLen, result := serial.Exchange(&transportRequest, requestLen, &transportResponse)
	return result == serial.ResultOK && bytesMatchLiteral(&transportResponse, responseLen, `{"health":"ok"}`)
}

func serialBridgePlan(request agent.BridgeRequest) agent.BridgeResult {
	requestLen, ok := encodeBridgeRequest(request, &transportRequest)
	if !ok {
		return agent.BridgeResult{OK: false, Reason: agent.MessageLLMBridgeFailed}
	}
	responseLen, result := serial.Exchange(&transportRequest, requestLen, &transportResponse)
	if result != serial.ResultOK {
		reason := agent.MessageLLMBridgeFailed
		if result == serial.ResultWriteTimeout || result == serial.ResultReadTimeout || result == serial.ResultPartialFrame {
			reason = agent.MessageBridgeTimeout
		}
		return agent.BridgeResult{OK: false, Reason: reason}
	}
	response, ok := decodeBridgeResponse(&transportResponse, responseLen)
	if !ok {
		return agent.BridgeResult{OK: false, Reason: agent.MessageLLMBridgeFailed}
	}
	return agent.BridgeResult{OK: true, Response: response}
}

type bridgeJSONWriter struct {
	buffer *[serial.MaxPayload]byte
	length int
	ok     bool
}

func encodeBridgeRequest(request agent.BridgeRequest, output *[serial.MaxPayload]byte) (int, bool) {
	writer := bridgeJSONWriter{buffer: output, ok: output != nil}
	bridgeWriteLiteral(&writer, `{"input":"`)
	writer.escaped(request.Input[:], request.InputLen)
	bridgeWriteLiteral(&writer, `","context":`)
	if request.HasContext {
		bridgeWriteLiteral(&writer, `{"lastIntent":"`)
		writer.intentName(request.Context.LastIntent)
		bridgeWriteLiteral(&writer, `","lastAction":"`)
		writer.actionName(request.Context.LastAction)
		bridgeWriteLiteral(&writer, `","lastSummary":"`)
		writer.messageName(request.Context.LastResultSummary)
		bridgeWriteLiteral(&writer, `","requestCount":`)
		writer.uint(request.Context.RequestCount)
		writer.byte('}')
	} else {
		bridgeWriteLiteral(&writer, `{}`)
	}
	bridgeWriteLiteral(&writer, `,"allowedActions":[`)
	if request.AllowedActionCount < 0 || request.AllowedActionCount > agent.MaxBridgeAllowedActions {
		writer.ok = false
	}
	for i := 0; writer.ok && i < request.AllowedActionCount; i++ {
		if i > 0 {
			writer.byte(',')
		}
		writer.byte('"')
		if !writer.actionName(request.AllowedActions[i]) {
			writer.ok = false
			break
		}
		writer.byte('"')
	}
	bridgeWriteLiteral(&writer, `]}`)
	return writer.length, writer.ok
}

func (w *bridgeJSONWriter) byte(value byte) {
	if !w.ok || w.length >= serial.MaxRequestPayload {
		w.ok = false
		return
	}
	w.buffer[w.length] = value
	w.length++
}

func bridgeWriteLiteral(w *bridgeJSONWriter, value string) {
	for i := 0; i < len(value); i++ {
		w.byte(value[i])
	}
}

func (w *bridgeJSONWriter) escaped(value []byte, length int) {
	if length < 0 || length > len(value) {
		w.ok = false
		return
	}
	const hex = "0123456789abcdef"
	for i := 0; i < length; i++ {
		current := value[i]
		switch current {
		case '"', '\\':
			w.byte('\\')
			w.byte(current)
		case '\b':
			bridgeWriteLiteral(w, `\b`)
		case '\f':
			bridgeWriteLiteral(w, `\f`)
		case '\n':
			bridgeWriteLiteral(w, `\n`)
		case '\r':
			bridgeWriteLiteral(w, `\r`)
		case '\t':
			bridgeWriteLiteral(w, `\t`)
		default:
			if current < 0x20 {
				bridgeWriteLiteral(w, `\u00`)
				w.byte(hex[current>>4])
				w.byte(hex[current&0x0f])
			} else {
				w.byte(current)
			}
		}
	}
}

func (w *bridgeJSONWriter) uint(value uint64) {
	if value == 0 {
		w.byte('0')
		return
	}
	var digits [20]byte
	count := 0
	for value > 0 {
		digits[count] = byte(value%10) + '0'
		value /= 10
		count++
	}
	for i := count - 1; i >= 0; i-- {
		w.byte(digits[i])
	}
}

func (w *bridgeJSONWriter) actionName(kind agent.ActionKind) bool {
	switch kind {
	case agent.ActionListFiles:
		bridgeWriteLiteral(w, "list_files")
	case agent.ActionReadFile:
		bridgeWriteLiteral(w, "read_file")
	case agent.ActionWriteFile:
		bridgeWriteLiteral(w, "write_file")
	case agent.ActionDeleteFile:
		bridgeWriteLiteral(w, "delete_file")
	case agent.ActionStatFile:
		bridgeWriteLiteral(w, "stat_file")
	case agent.ActionShowHelp:
		bridgeWriteLiteral(w, "show_help")
	case agent.ActionShowHistory:
		bridgeWriteLiteral(w, "show_history")
	case agent.ActionShowVersion:
		bridgeWriteLiteral(w, "show_version")
	case agent.ActionShowTicks:
		bridgeWriteLiteral(w, "show_ticks")
	case agent.ActionShowMemoryMap:
		bridgeWriteLiteral(w, "show_memory_map")
	case agent.ActionSetMode:
		bridgeWriteLiteral(w, "set_mode")
	default:
		bridgeWriteLiteral(w, "unknown")
		return false
	}
	return true
}

func (w *bridgeJSONWriter) intentName(kind agent.IntentKind) {
	w.actionName(agent.ActionKind(kind))
}

func (w *bridgeJSONWriter) messageName(kind agent.MessageKind) {
	switch kind {
	case agent.MessageNone:
		bridgeWriteLiteral(w, "none")
	case agent.MessageOK:
		bridgeWriteLiteral(w, "ok")
	case agent.MessageFilesListed:
		bridgeWriteLiteral(w, "files_listed")
	case agent.MessageNoFiles:
		bridgeWriteLiteral(w, "no_files")
	case agent.MessageFileRead:
		bridgeWriteLiteral(w, "file_read")
	case agent.MessageFileStat:
		bridgeWriteLiteral(w, "file_stat")
	case agent.MessageFileNotFound:
		bridgeWriteLiteral(w, "file_not_found")
	case agent.MessageHistoryListed:
		bridgeWriteLiteral(w, "history_listed")
	case agent.MessageVersionShown:
		bridgeWriteLiteral(w, "version_shown")
	case agent.MessageTicksShown:
		bridgeWriteLiteral(w, "ticks_shown")
	case agent.MessageMemoryMapShown:
		bridgeWriteLiteral(w, "memory_map_shown")
	default:
		bridgeWriteLiteral(w, "other")
	}
}

type bridgeJSONReader struct {
	buffer *[serial.MaxPayload]byte
	length int
	offset int
}

func decodeBridgeResponse(input *[serial.MaxPayload]byte, inputLen int) (agent.BridgeResponse, bool) {
	var response agent.BridgeResponse
	if input == nil || inputLen < 0 || inputLen > serial.MaxResponsePayload {
		return response, false
	}
	reader := bridgeJSONReader{buffer: input, length: inputLen}
	if !bridgeReaderLiteral(&reader, `{"`) {
		return response, false
	}
	if bridgeReaderPeekLiteral(&reader, `error"`) {
		return response, false
	}
	if !bridgeReaderLiteral(&reader, `action":"`) {
		return response, false
	}
	var actionName [32]byte
	actionLen, ok := reader.string(&actionName)
	if !ok || !bridgeReaderLiteral(&reader, `,"args":[`) {
		return response, false
	}
	response.Action, ok = bridgeActionFromBytes(&actionName, actionLen)
	if !ok {
		return response, false
	}
	if reader.consume(']') {
		response.TargetLen = 0
	} else {
		if !reader.consume('"') {
			return response, false
		}
		var target [32]byte
		targetLen, targetOK := reader.string(&target)
		if !targetOK || targetLen == 0 || targetLen > agent.MaxNameLen || !bridgeReaderLiteral(&reader, `]`) {
			return response, false
		}
		response.TargetLen = targetLen
		for i := 0; i < targetLen; i++ {
			response.Target[i] = target[i]
		}
	}
	if !reader.consume(',') {
		return response, false
	}
	if bridgeReaderPeekLiteral(&reader, `"explanation":"`) {
		if !bridgeReaderLiteral(&reader, `"explanation":"`) || !reader.skipString() || !reader.consume(',') {
			return response, false
		}
	}
	if !bridgeReaderLiteral(&reader, `"intent":"`) {
		return response, false
	}
	var intentName [32]byte
	intentLen, ok := reader.string(&intentName)
	if !ok || !bridgeReaderLiteral(&reader, `,"risk":"`) {
		return response, false
	}
	response.Intent, ok = bridgeIntentFromBytes(&intentName, intentLen)
	if !ok {
		return response, false
	}
	var riskName [32]byte
	riskLen, ok := reader.string(&riskName)
	if !ok || !reader.consume('}') || reader.offset != reader.length {
		return response, false
	}
	response.Risk, ok = bridgeRiskFromBytes(&riskName, riskLen)
	return response, ok
}

func (r *bridgeJSONReader) consume(value byte) bool {
	if r.offset >= r.length || r.buffer[r.offset] != value {
		return false
	}
	r.offset++
	return true
}

func bridgeReaderLiteral(r *bridgeJSONReader, value string) bool {
	if !bridgeReaderPeekLiteral(r, value) {
		return false
	}
	r.offset += len(value)
	return true
}

func bridgeReaderPeekLiteral(r *bridgeJSONReader, value string) bool {
	if r.offset+len(value) > r.length {
		return false
	}
	for i := 0; i < len(value); i++ {
		if r.buffer[r.offset+i] != value[i] {
			return false
		}
	}
	return true
}

func (r *bridgeJSONReader) string(output *[32]byte) (int, bool) {
	length := 0
	for r.offset < r.length {
		value := r.buffer[r.offset]
		r.offset++
		if value == '"' {
			return length, true
		}
		if value < 0x20 {
			return 0, false
		}
		if value != '\\' {
			if !appendBridgeByte(output, &length, value) {
				return 0, false
			}
			continue
		}
		if r.offset >= r.length {
			return 0, false
		}
		escape := r.buffer[r.offset]
		r.offset++
		switch escape {
		case '"', '\\', '/':
			if !appendBridgeByte(output, &length, escape) {
				return 0, false
			}
		case 'u':
			codePoint, ok := r.hexCodeUnit()
			if !ok {
				return 0, false
			}
			if codePoint >= 0xd800 && codePoint <= 0xdbff {
				if !r.consume('\\') || !r.consume('u') {
					return 0, false
				}
				low, lowOK := r.hexCodeUnit()
				if !lowOK || low < 0xdc00 || low > 0xdfff {
					return 0, false
				}
				codePoint = 0x10000 + (codePoint-0xd800)<<10 + low - 0xdc00
			} else if codePoint >= 0xdc00 && codePoint <= 0xdfff {
				return 0, false
			}
			if !appendBridgeCodePoint(output, &length, codePoint) {
				return 0, false
			}
		default:
			return 0, false
		}
	}
	return 0, false
}

func (r *bridgeJSONReader) hexCodeUnit() (uint32, bool) {
	var value uint32
	for i := 0; i < 4; i++ {
		if r.offset >= r.length {
			return 0, false
		}
		digit, ok := hexDigitValue(r.buffer[r.offset])
		if !ok {
			return 0, false
		}
		r.offset++
		value = value<<4 | uint32(digit)
	}
	return value, true
}

func appendBridgeCodePoint(output *[32]byte, length *int, value uint32) bool {
	if value <= 0x7f {
		return value >= 0x20 && appendBridgeByte(output, length, byte(value))
	}
	if value <= 0x7ff {
		return appendBridgeByte(output, length, byte(0xc0|value>>6)) &&
			appendBridgeByte(output, length, byte(0x80|value&0x3f))
	}
	if value <= 0xffff {
		return appendBridgeByte(output, length, byte(0xe0|value>>12)) &&
			appendBridgeByte(output, length, byte(0x80|value>>6&0x3f)) &&
			appendBridgeByte(output, length, byte(0x80|value&0x3f))
	}
	if value <= 0x10ffff {
		return appendBridgeByte(output, length, byte(0xf0|value>>18)) &&
			appendBridgeByte(output, length, byte(0x80|value>>12&0x3f)) &&
			appendBridgeByte(output, length, byte(0x80|value>>6&0x3f)) &&
			appendBridgeByte(output, length, byte(0x80|value&0x3f))
	}
	return false
}

func appendBridgeByte(output *[32]byte, length *int, value byte) bool {
	if *length >= len(output) {
		return false
	}
	output[*length] = value
	(*length)++
	return true
}

func (r *bridgeJSONReader) skipString() bool {
	for r.offset < r.length {
		value := r.buffer[r.offset]
		r.offset++
		if value == '"' {
			return true
		}
		if value < 0x20 {
			return false
		}
		if value != '\\' {
			continue
		}
		if r.offset >= r.length {
			return false
		}
		escape := r.buffer[r.offset]
		r.offset++
		if escape == 'u' {
			for i := 0; i < 4; i++ {
				if r.offset >= r.length || !isHexDigit(r.buffer[r.offset]) {
					return false
				}
				r.offset++
			}
		} else if escape != '"' && escape != '\\' && escape != '/' && escape != 'b' && escape != 'f' && escape != 'n' && escape != 'r' && escape != 't' {
			return false
		}
	}
	return false
}

func isHexDigit(value byte) bool {
	_, ok := hexDigitValue(value)
	return ok
}

func hexDigitValue(value byte) (byte, bool) {
	if value >= '0' && value <= '9' {
		return value - '0', true
	}
	if value >= 'a' && value <= 'f' {
		return value - 'a' + 10, true
	}
	if value >= 'A' && value <= 'F' {
		return value - 'A' + 10, true
	}
	return 0, false
}

func bridgeActionFromBytes(value *[32]byte, length int) (agent.ActionKind, bool) {
	if bytesMatchLiteral32(value, length, "list_files") {
		return agent.ActionListFiles, true
	}
	if bytesMatchLiteral32(value, length, "read_file") {
		return agent.ActionReadFile, true
	}
	if bytesMatchLiteral32(value, length, "write_file") {
		return agent.ActionWriteFile, true
	}
	if bytesMatchLiteral32(value, length, "delete_file") {
		return agent.ActionDeleteFile, true
	}
	if bytesMatchLiteral32(value, length, "stat_file") {
		return agent.ActionStatFile, true
	}
	if bytesMatchLiteral32(value, length, "show_help") {
		return agent.ActionShowHelp, true
	}
	if bytesMatchLiteral32(value, length, "show_history") {
		return agent.ActionShowHistory, true
	}
	if bytesMatchLiteral32(value, length, "show_version") {
		return agent.ActionShowVersion, true
	}
	if bytesMatchLiteral32(value, length, "show_ticks") {
		return agent.ActionShowTicks, true
	}
	if bytesMatchLiteral32(value, length, "show_memory_map") {
		return agent.ActionShowMemoryMap, true
	}
	if bytesMatchLiteral32(value, length, "set_mode") {
		return agent.ActionSetMode, true
	}
	return agent.ActionUnknown, false
}

func bridgeIntentFromBytes(value *[32]byte, length int) (agent.IntentKind, bool) {
	action, ok := bridgeActionFromBytes(value, length)
	return agent.IntentKind(action), ok
}

func bridgeRiskFromBytes(value *[32]byte, length int) (agent.RiskLevel, bool) {
	if bytesMatchLiteral32(value, length, "safe") {
		return agent.RiskSafe, true
	}
	if bytesMatchLiteral32(value, length, "risky") {
		return agent.RiskRisky, true
	}
	return agent.RiskSafe, false
}

func copyLiteral(output *[serial.MaxPayload]byte, value string) int {
	for i := 0; i < len(value); i++ {
		output[i] = value[i]
	}
	return len(value)
}

func bytesMatchLiteral(value *[serial.MaxPayload]byte, length int, literal string) bool {
	if value == nil || length != len(literal) {
		return false
	}
	for i := 0; i < length; i++ {
		if value[i] != literal[i] {
			return false
		}
	}
	return true
}

func bytesMatchLiteral32(value *[32]byte, length int, literal string) bool {
	if value == nil || length != len(literal) {
		return false
	}
	for i := 0; i < length; i++ {
		if value[i] != literal[i] {
			return false
		}
	}
	return true
}
