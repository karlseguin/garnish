package garnish

import (
	"bytes"
	"fmt"
	"github.com/karlseguin/garnish/core"
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
	spec.Expect(buffer.String()).ToEqual("[internal] [internal] info msg\n")
}

func TestLoggerIncludesTheRequestIdForInfoMessage(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Infof(core.ContextBuilder().SetId("4994"), "info msg")
	spec.Expect(buffer.String()).ToEqual("[4994] [cb] info msg\n")
}

func TestLoggerLogsErrorsWhenInfoIsOff(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(false)
	l.Errorf(nil, "error msg1")
	spec.Expect(buffer.String()).ToEqual("[internal] [internal] error msg1\n")
}

func TestLoggerLogsErrorsWhenInfoIsOn(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Errorf(nil, "error msg2")
	spec.Expect(buffer.String()).ToEqual("[internal] [internal] error msg2\n")
}

func TestLoggerIncludesTheRequestIdForErrorMessage(t *testing.T) {
	spec := gspec.New(t)
	l, buffer := testLogger(true)
	l.Errorf(core.ContextBuilder().SetId("8664"), "error msg")
	spec.Expect(buffer.String()).ToEqual("[8664] [cb] error msg\n")
}

func testLogger(info bool) (core.Logger, *bytes.Buffer) {
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

func (l *FakeLogger) Infof(context core.Context, format string, v ...interface{}) {
	l.printf(true, format, v...)
}

func (l *FakeLogger) Info(context core.Context, v ...interface{}) {
	l.print(true, v...)
}

func (l *FakeLogger) Errorf(context core.Context, format string, v ...interface{}) {
	l.printf(false, format, v...)
}

func (l *FakeLogger) Error(context core.Context, v ...interface{}) {
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
