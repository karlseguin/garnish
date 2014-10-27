package gc

import (
	"os"
	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("garnish")

func init() {
	var format = "[%{level:.4s}] [%{time:Jan _2 15:04:05.000}] %{message}"
	logging.SetLevel(logging.WARNING, "garnish")
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
	logging.SetFormatter(logging.MustStringFormatter(format))
}
