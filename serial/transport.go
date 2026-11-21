package serial

const maxPortPolls = 1000000

var (
	encodedRequest  [MaxFrameSize]byte
	responseDecoder Decoder
)

// Exchange sends one request and waits for the first response.
func Exchange(request *[MaxPayload]byte, requestLen int, response *[MaxPayload]byte) (int, Result) {
	return ExchangeUntil(request, requestLen, response, nil)
}

// ExchangeUntil shares one receive budget across all frames. Like Exchange,
// it is single-flight so the freestanding guest can use fixed-size storage.
func ExchangeUntil(request *[MaxPayload]byte, requestLen int, response *[MaxPayload]byte, accept func(*[MaxPayload]byte, int) bool) (int, Result) {
	if response == nil {
		return 0, ResultInvalidPayload
	}
	frameLen, result := Encode(FrameRequest, request, requestLen, &encodedRequest)
	if result != ResultOK {
		return 0, result
	}
	if !initPort() {
		return 0, ResultPortUnavailable
	}

	drainInput()
	written := 0
	for polls := 0; written < frameLen && polls < maxPortPolls; polls++ {
		if tryWriteByte(encodedRequest[written]) {
			written++
		}
	}
	if written != frameLen {
		return 0, ResultWriteTimeout
	}

	responseDecoder.Reset()
	bytesRead := 0
	for polls := 0; polls < maxPortPolls; polls++ {
		value, ok := tryReadByte()
		if !ok {
			continue
		}
		bytesRead++
		result = responseDecoder.Push(value)
		if result == ResultNeedMore {
			continue
		}
		if result != ResultOK {
			return 0, result
		}

		frame := responseDecoder.Frame()
		if frame.Kind != FrameResponse {
			return 0, ResultUnexpectedFrame
		}
		for i := 0; i < frame.Length; i++ {
			response[i] = frame.Payload[i]
		}
		if accept == nil || accept(response, frame.Length) {
			return frame.Length, ResultOK
		}
		responseDecoder.Reset()
		bytesRead = 0
	}
	if bytesRead != 0 {
		return 0, ResultPartialFrame
	}
	return 0, ResultReadTimeout
}

func drainInput() {
	for i := 0; i < MaxFrameSize*2; i++ {
		if _, ok := tryReadByte(); !ok {
			return
		}
	}
}
