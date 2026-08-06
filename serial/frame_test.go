package serial

import "testing"

func TestChecksumMatchesCCITTFalseVector(t *testing.T) {
	checksum := uint16(0xFFFF)
	for i := 0; i < len("123456789"); i++ {
		checksum = updateChecksum(checksum, "123456789"[i])
	}
	if checksum != 0x29B1 {
		t.Fatalf("checksum = 0x%04X, expected 0x29B1", checksum)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		kind    FrameKind
		payload string
	}{
		{name: "request", kind: FrameRequest, payload: "show me the files"},
		{name: "response", kind: FrameResponse, payload: "list_files"},
		{name: "empty response", kind: FrameResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := payloadForTest(tt.payload)
			var encoded [MaxFrameSize]byte
			encodedLen, result := Encode(tt.kind, &payload, len(tt.payload), &encoded)
			if result != ResultOK {
				t.Fatalf("Encode() result = %d", result)
			}

			var decoder Decoder
			decoder.Reset()
			for i := 0; i < encodedLen; i++ {
				result = decoder.Push(encoded[i])
				if i < encodedLen-1 && result != ResultNeedMore {
					t.Fatalf("Push() at %d = %d", i, result)
				}
			}
			if result != ResultOK {
				t.Fatalf("final Push() result = %d", result)
			}
			frame := decoder.Frame()
			if frame.Kind != tt.kind || frame.Length != len(tt.payload) {
				t.Fatalf("decoded frame = kind %d length %d", frame.Kind, frame.Length)
			}
			for i := 0; i < frame.Length; i++ {
				if frame.Payload[i] != tt.payload[i] {
					t.Fatalf("payload byte %d = %q", i, frame.Payload[i])
				}
			}
		})
	}
}

func TestFramingRejectsInvalidInput(t *testing.T) {
	var payload [MaxPayload]byte
	var output [MaxFrameSize]byte

	tests := []struct {
		name       string
		kind       FrameKind
		payloadLen int
		payload    *[MaxPayload]byte
		output     *[MaxFrameSize]byte
		want       Result
	}{
		{name: "negative length", kind: FrameRequest, payloadLen: -1, payload: &payload, output: &output, want: ResultInvalidPayload},
		{name: "nil payload", kind: FrameRequest, payloadLen: 1, output: &output, want: ResultInvalidPayload},
		{name: "nil output", kind: FrameRequest, payload: &payload, want: ResultInvalidPayload},
		{name: "unknown kind", kind: FrameKind(99), payload: &payload, output: &output, want: ResultMalformedFrame},
		{name: "request too large", kind: FrameRequest, payloadLen: MaxRequestPayload + 1, payload: &payload, output: &output, want: ResultOversizedFrame},
		{name: "response too large", kind: FrameResponse, payloadLen: MaxResponsePayload + 1, payload: &payload, output: &output, want: ResultOversizedFrame},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := Encode(tt.kind, tt.payload, tt.payloadLen, tt.output); got != tt.want {
				t.Fatalf("Encode() result = %d, expected %d", got, tt.want)
			}
		})
	}
}

func TestDecoderRejectsMalformedFrames(t *testing.T) {
	valid, validLen := encodedForTest(t, FrameResponse, "pong")

	tests := []struct {
		name  string
		frame [MaxFrameSize]byte
		len   int
		want  Result
	}{
		{name: "bad magic", frame: changedByte(valid, 0, 'X'), len: validLen, want: ResultMalformedFrame},
		{name: "bad version", frame: changedByte(valid, 2, FrameVersion+1), len: validLen, want: ResultMalformedFrame},
		{name: "bad kind", frame: changedByte(valid, 3, 99), len: validLen, want: ResultMalformedFrame},
		{name: "bad checksum", frame: changedByte(valid, validLen-1, valid[validLen-1]^0xFF), len: validLen, want: ResultMalformedFrame},
		{name: "partial", frame: valid, len: validLen - 1, want: ResultNeedMore},
		{name: "oversized response", frame: oversizedResponseHeader(), len: frameHeaderSize, want: ResultOversizedFrame},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoder Decoder
			decoder.Reset()
			result := ResultNeedMore
			for i := 0; i < tt.len; i++ {
				result = decoder.Push(tt.frame[i])
				if result != ResultNeedMore {
					break
				}
			}
			if result != tt.want {
				t.Fatalf("decoder result = %d, expected %d", result, tt.want)
			}
		})
	}
}

func TestExchangeSendsRequestAndReceivesResponse(t *testing.T) {
	reply, replyLen := encodedForTest(t, FrameResponse, "pong")
	ConfigureResponseForTesting(&reply, replyLen)
	t.Cleanup(ResetForTesting)

	request := payloadForTest("ping")
	var response [MaxPayload]byte
	responseLen, result := Exchange(&request, 4, &response)
	if result != ResultOK || responseLen != 4 || string(response[:responseLen]) != "pong" {
		t.Fatalf("Exchange() = length %d result %d payload %q", responseLen, result, response[:responseLen])
	}

	var decoder Decoder
	decoder.Reset()
	for i := 0; i < testOutputLen; i++ {
		result = decoder.Push(testOutput[i])
	}
	frame := decoder.Frame()
	if result != ResultOK || frame.Kind != FrameRequest || frame.Length != 4 || string(frame.Payload[:frame.Length]) != "ping" {
		t.Fatalf("outbound frame = kind %d length %d result %d", frame.Kind, frame.Length, result)
	}
}

func TestExchangeFailsSafely(t *testing.T) {
	valid, validLen := encodedForTest(t, FrameResponse, "pong")
	request := payloadForTest("ping")
	var response [MaxPayload]byte

	tests := []struct {
		name      string
		configure func()
		want      Result
	}{
		{name: "port unavailable", configure: ResetForTesting, want: ResultPortUnavailable},
		{name: "write timeout", configure: func() {
			ResetForTesting()
			testPortAvailable = true
		}, want: ResultWriteTimeout},
		{name: "read timeout", configure: func() {
			ConfigureResponseForTesting(nil, 0)
		}, want: ResultReadTimeout},
		{name: "partial frame", configure: func() {
			ConfigureResponseForTesting(&valid, validLen-1)
		}, want: ResultPartialFrame},
		{name: "bad checksum", configure: func() {
			malformed := changedByte(valid, validLen-1, valid[validLen-1]^0xFF)
			ConfigureResponseForTesting(&malformed, validLen)
		}, want: ResultMalformedFrame},
		{name: "oversized response", configure: func() {
			oversized := oversizedResponseHeader()
			ConfigureResponseForTesting(&oversized, frameHeaderSize)
		}, want: ResultOversizedFrame},
		{name: "request returned as response", configure: func() {
			wrongKind, wrongKindLen := encodedForTest(t, FrameRequest, "pong")
			ConfigureResponseForTesting(&wrongKind, wrongKindLen)
		}, want: ResultUnexpectedFrame},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.configure()
			if _, result := Exchange(&request, 4, &response); result != tt.want {
				t.Fatalf("Exchange() result = %d, expected %d", result, tt.want)
			}
		})
	}
	ResetForTesting()
}

func TestExchangeRecoversAfterMalformedResponse(t *testing.T) {
	valid, validLen := encodedForTest(t, FrameResponse, "pong")
	malformed := changedByte(valid, validLen-1, valid[validLen-1]^0xFF)
	request := payloadForTest("ping")
	var response [MaxPayload]byte

	ConfigureResponseForTesting(&malformed, validLen)
	if _, result := Exchange(&request, 4, &response); result != ResultMalformedFrame {
		t.Fatalf("malformed Exchange() result = %d", result)
	}
	ConfigureResponseForTesting(&valid, validLen)
	if responseLen, result := Exchange(&request, 4, &response); result != ResultOK || responseLen != 4 {
		t.Fatalf("recovery Exchange() = length %d result %d", responseLen, result)
	}
	ResetForTesting()
}

func payloadForTest(value string) [MaxPayload]byte {
	var payload [MaxPayload]byte
	for i := 0; i < len(value); i++ {
		payload[i] = value[i]
	}
	return payload
}

func encodedForTest(t *testing.T, kind FrameKind, value string) ([MaxFrameSize]byte, int) {
	t.Helper()
	payload := payloadForTest(value)
	var encoded [MaxFrameSize]byte
	encodedLen, result := Encode(kind, &payload, len(value), &encoded)
	if result != ResultOK {
		t.Fatalf("Encode() result = %d", result)
	}
	return encoded, encodedLen
}

func changedByte(frame [MaxFrameSize]byte, offset int, value byte) [MaxFrameSize]byte {
	frame[offset] = value
	return frame
}

func oversizedResponseHeader() [MaxFrameSize]byte {
	var frame [MaxFrameSize]byte
	frame[0] = frameMagicFirst
	frame[1] = frameMagicSecond
	frame[2] = FrameVersion
	frame[3] = byte(FrameResponse)
	length := MaxResponsePayload + 1
	frame[4] = byte(length >> 8)
	frame[5] = byte(length)
	return frame
}
