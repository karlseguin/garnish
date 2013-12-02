package garnish

import (
	"bytes"
	"fmt"
	"github.com/karlseguin/gspec"
	"log"
	"testing"
)

func TestLoggerDoesNotLogInfoMessagesWhenInfoIsOff(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(false)
	l.Infof(nil, "info msg")
	spec.Expect(buffer.Len()).ToEqual(0)
}

func TestLoggerLogsInfoMessagesWhenInfoIsOn(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Infof(nil, "info msg")
	spec.Expect(buffer.String()).ToEqual("[internal] info msg\n")
}

func TestLoggerIncludesTheRequestIdForInfoMessage(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Infof(ContextBuilder().SetId("4994"), "info msg")
	spec.Expect(buffer.String()).ToEqual("[4994] info msg\n")
}

func TestLoggerLogsErrorsWhenInfoIsOff(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(false)
	l.Errorf(nil, "error msg1")
	spec.Expect(buffer.String()).ToEqual("[internal] error msg1\n")
}

func TestLoggerLogsErrorsWhenInfoIsOn(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Errorf(nil, "error msg2")
	spec.Expect(buffer.String()).ToEqual("[internal] error msg2\n")
}

func TestLoggerIncludesTheRequestIdForErrorMessage(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Errorf(ContextBuilder().SetId("8664"), "error msg")
	spec.Expect(buffer.String()).ToEqual("[8664] error msg\n")
}

func testLogger(info bool) (Logger, *bytes.Buffer) {
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

func (l *FakeLogger) Infof(context Context, format string, v ...interface{}) {
	l.printf(true, format, v...)
}

func (l *FakeLogger) Info(context Context, v ...interface{}) {
	l.print(true, v...)
}

func (l *FakeLogger) Errorf(context Context, format string, v ...interface{}) {
	l.printf(false, format, v...)
}

func (l *FakeLogger) Error(context Context, v ...interface{}) {
	l.print(false, v...)
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
