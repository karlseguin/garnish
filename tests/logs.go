package tests

import (
	. "github.com/karlseguin/expect"
	"github.com/op/go-logging"
)

var Logs = logging.InitForTesting(logging.WARNING)

func ExpectError(expected string) {
	for node := Logs.Head(); node != nil; node = node.Next() {
		if node.Record.Message() == expected {
			return
		}
	}
	Fail("expected %q to have been logged, last: %q", expected, Logs.Head().Record.Message())
}
