package serial

const (
	FrameVersion byte = 1

	MaxRequestPayload  = 2048
	MaxResponsePayload = 1024
	MaxPayload         = MaxRequestPayload

	frameHeaderSize  = 6
	frameTrailerSize = 2
	MaxFrameSize     = frameHeaderSize + MaxPayload + frameTrailerSize
)

const (
	frameMagicFirst  byte = 'D'
	frameMagicSecond byte = 'V'
)

type FrameKind uint8

const (
	FrameRequest FrameKind = iota + 1
	FrameResponse
)

type Result uint8

const (
	ResultOK Result = iota
	ResultNeedMore
	ResultInvalidPayload
	ResultPortUnavailable
	ResultWriteTimeout
	ResultReadTimeout
	ResultPartialFrame
	ResultMalformedFrame
	ResultOversizedFrame
	ResultUnexpectedFrame
)

type Frame struct {
	Kind    FrameKind
	Payload [MaxPayload]byte
	Length  int
}

type Decoder struct {
	frame       Frame
	state       uint8
	checksum    uint16
	receivedCRC uint16
	payloadRead int
}

const (
	decodeMagicFirst uint8 = iota
	decodeMagicSecond
	decodeVersion
	decodeKind
	decodeLengthHigh
	decodeLengthLow
	decodePayload
	decodeChecksumHigh
	decodeChecksumLow
)

func Encode(kind FrameKind, payload *[MaxPayload]byte, payloadLen int, output *[MaxFrameSize]byte) (int, Result) {
	if output == nil || payloadLen < 0 || payload == nil && payloadLen != 0 {
		return 0, ResultInvalidPayload
	}
	if !kind.valid() {
		return 0, ResultMalformedFrame
	}
	if payloadLen > MaxPayload || payloadLen > kind.maxPayload() {
		return 0, ResultOversizedFrame
	}

	output[0] = frameMagicFirst
	output[1] = frameMagicSecond
	output[2] = FrameVersion
	output[3] = byte(kind)
	output[4] = byte(payloadLen >> 8)
	output[5] = byte(payloadLen)

	checksum := uint16(0xFFFF)
	for i := 2; i < frameHeaderSize; i++ {
		checksum = updateChecksum(checksum, output[i])
	}
	for i := 0; i < payloadLen; i++ {
		output[frameHeaderSize+i] = payload[i]
		checksum = updateChecksum(checksum, payload[i])
	}

	checksumOffset := frameHeaderSize + payloadLen
	output[checksumOffset] = byte(checksum >> 8)
	output[checksumOffset+1] = byte(checksum)
	return checksumOffset + frameTrailerSize, ResultOK
}

func (d *Decoder) Reset() {
	d.frame.Kind = 0
	d.frame.Length = 0
	d.state = decodeMagicFirst
	d.checksum = 0xFFFF
	d.receivedCRC = 0
	d.payloadRead = 0
}

func (d *Decoder) Push(value byte) Result {
	switch d.state {
	case decodeMagicFirst:
		if value != frameMagicFirst {
			d.Reset()
			return ResultMalformedFrame
		}
		d.state = decodeMagicSecond
	case decodeMagicSecond:
		if value != frameMagicSecond {
			d.Reset()
			return ResultMalformedFrame
		}
		d.state = decodeVersion
	case decodeVersion:
		if value != FrameVersion {
			d.Reset()
			return ResultMalformedFrame
		}
		d.checksum = updateChecksum(d.checksum, value)
		d.state = decodeKind
	case decodeKind:
		d.frame.Kind = FrameKind(value)
		if !d.frame.Kind.valid() {
			d.Reset()
			return ResultMalformedFrame
		}
		d.checksum = updateChecksum(d.checksum, value)
		d.state = decodeLengthHigh
	case decodeLengthHigh:
		d.frame.Length = int(value) << 8
		d.checksum = updateChecksum(d.checksum, value)
		d.state = decodeLengthLow
	case decodeLengthLow:
		d.frame.Length |= int(value)
		d.checksum = updateChecksum(d.checksum, value)
		if d.frame.Length > d.frame.Kind.maxPayload() {
			d.Reset()
			return ResultOversizedFrame
		}
		if d.frame.Length == 0 {
			d.state = decodeChecksumHigh
		} else {
			d.state = decodePayload
		}
	case decodePayload:
		d.frame.Payload[d.payloadRead] = value
		d.payloadRead++
		d.checksum = updateChecksum(d.checksum, value)
		if d.payloadRead == d.frame.Length {
			d.state = decodeChecksumHigh
		}
	case decodeChecksumHigh:
		d.receivedCRC = uint16(value) << 8
		d.state = decodeChecksumLow
	case decodeChecksumLow:
		d.receivedCRC |= uint16(value)
		if d.receivedCRC != d.checksum {
			d.Reset()
			return ResultMalformedFrame
		}
		return ResultOK
	default:
		d.Reset()
		return ResultMalformedFrame
	}
	return ResultNeedMore
}

func (d *Decoder) Frame() *Frame {
	return &d.frame
}

func (kind FrameKind) valid() bool {
	return kind == FrameRequest || kind == FrameResponse
}

func (kind FrameKind) maxPayload() int {
	if kind == FrameResponse {
		return MaxResponsePayload
	}
	return MaxRequestPayload
}

func updateChecksum(checksum uint16, value byte) uint16 {
	checksum ^= uint16(value) << 8
	for i := 0; i < 8; i++ {
		if checksum&0x8000 != 0 {
			checksum = checksum<<1 ^ 0x1021
		} else {
			checksum <<= 1
		}
	}
	return checksum
}
