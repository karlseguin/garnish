package garnish

import (
	"fmt"
	. "github.com/karlseguin/expect"
	"testing"
)

type ConfigurationTests struct{}

func Test_Configuration(t *testing.T) {
	Expectify(new(ConfigurationTests), t)
}

func (_ ConfigurationTests) FailedBuildWithoutUpstream() {
	logger := NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.errors).To.Contain("Atleast one upstream must be configured")
}

func (_ ConfigurationTests) FailedBuildWithMissingUpstreamAddress() {
	logger := NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	c.Upstream("test")
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.errors).To.Contain(`Upstream test has an invalid address: ""`)
}

func (_ ConfigurationTests) FailedBuildWithInvalidUpstreamAddress() {
	logger := NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Upstream("test2").Address("128.93.202.0")
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.errors).To.Contain(`Upstream test2's address should begin with unix:/, http:// or https://`)
}

func (_ ConfigurationTests) FailedBuildWithoutRoute() {
	logger := NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	c.Upstream("test1").Address("http://openmymind.net/")
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.errors).To.Contain("Atleast one route must be configured")
}

type FakeLogger struct {
	infos  []string
	warns  []string
	errors []string
}

func NewFakeLogger() *FakeLogger {
	return &FakeLogger{}
}

func (f *FakeLogger) Info(message string) {
	f.infos = append(f.infos, message)
}

func (f *FakeLogger) Infof(format string, v ...interface{}) {
	f.infos = append(f.infos, fmt.Sprintf(format, v...))
}

func (f *FakeLogger) Warn(message string) {
	f.warns = append(f.warns, message)
}

func (f *FakeLogger) Warnf(format string, v ...interface{}) {
	f.warns = append(f.warns, fmt.Sprintf(format, v...))
}

func (f *FakeLogger) Error(message string) {
	f.errors = append(f.errors, message)
}

func (f *FakeLogger) Errorf(format string, v ...interface{}) {
	f.errors = append(f.errors, fmt.Sprintf(format, v...))
}

func (f *FakeLogger) Verbose() {}

func (f *FakeLogger) IsVerbose() bool {
	return false
}
