//go:build testing

package serial

var (
	testPortAvailable  bool
	testWriteAvailable bool
	testReply          [MaxFrameSize]byte
	testReplyLen       int
	testReplyQueued    bool
	testInput          [MaxFrameSize]byte
	testInputLen       int
	testInputOffset    int
	testOutput         [MaxFrameSize]byte
	testOutputLen      int
)

func ResetForTesting() {
	testPortAvailable = false
	testWriteAvailable = false
	testReplyLen = 0
	testReplyQueued = false
	testInputLen = 0
	testInputOffset = 0
	testOutputLen = 0
}

func ConfigureResponseForTesting(frame *[MaxFrameSize]byte, frameLen int) {
	ResetForTesting()
	testPortAvailable = true
	testWriteAvailable = true
	if frame == nil || frameLen < 0 || frameLen > MaxFrameSize {
		return
	}
	for i := 0; i < frameLen; i++ {
		testReply[i] = frame[i]
	}
	testReplyLen = frameLen
}

func initPort() bool {
	return testPortAvailable
}

func tryWriteByte(value byte) bool {
	if !testWriteAvailable || testOutputLen >= MaxFrameSize {
		return false
	}
	testOutput[testOutputLen] = value
	testOutputLen++
	if !testReplyQueued && testReplyLen > 0 {
		for i := 0; i < testReplyLen; i++ {
			testInput[i] = testReply[i]
		}
		testInputLen = testReplyLen
		testReplyQueued = true
	}
	return true
}

func tryReadByte() (byte, bool) {
	if testInputOffset >= testInputLen {
		return 0, false
	}
	value := testInput[testInputOffset]
	testInputOffset++
	return value, true
}
