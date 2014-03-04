package garnish

import (
	"bytes"
	"fmt"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/gspec"
	"log"
	"testing"
)

func TestLoggerDoesNotLogInfoMessagesWhenInfoIsOff(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(false)
	l.Infof("info msg")
	spec.Expect(buffer.Len()).ToEqual(0)
}

func TestLoggerLogsInfoMessagesWhenInfoIsOn(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Infof("info msg")
	spec.Expect(buffer.String()).ToEqual("info msg\n")
}

func TestLoggerLogsErrorsWhenInfoIsOn(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Errorf("error msg2")
	spec.Expect(buffer.String()).ToEqual("error msg2\n")
}

func TestLoggerIncludesTheRequestIdForErrorMessage(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Errorf("error msg")
	spec.Expect(buffer.String()).ToEqual("error msg\n")
}

func testLogger(info bool) (gc.Logger, *bytes.Buffer) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	return &logger{info, log.New(buffer, "", 0)}, buffer
}

type FakeLogger struct {
	messages []FakeLogMessage
}

type FakeLogMessage struct {
	info    bool
	message string
}

func (l *FakeLogger) Infof(format string, v ...interface{}) {
	l.printf(true, format, v...)
}

func (l *FakeLogger) Info(v ...interface{}) {
	l.print(true, v...)
}

func (l *FakeLogger) Errorf(format string, v ...interface{}) {
	l.printf(false, format, v...)
}

func (l *FakeLogger) Error(v ...interface{}) {
	l.print(false, v...)
}

func (l *FakeLogger) LogInfo() bool {
	return true
}

func (l *FakeLogger) printf(info bool, format string, v ...interface{}) {
	if l.messages == nil {
		l.messages = make([]FakeLogMessage, 0, 1)
	}
	l.messages = append(l.messages, FakeLogMessage{info, fmt.Sprintf(format, v...)})
}

func (l *FakeLogger) print(info bool, v ...interface{}) {
	if l.messages == nil {
		l.messages = make([]FakeLogMessage, 0, 1)
	}
	l.messages = append(l.messages, FakeLogMessage{info, fmt.Sprint(v...)})
}

func (l *FakeLogger) Assert(t *testing.T, expected FakeLogMessage) {
	for _, actual := range l.messages {
		if actual == expected {
			return
		}
	}
	t.Errorf("expected logged message %v", expected)
}
