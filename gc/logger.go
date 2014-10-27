package gc

import (
	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("garnish")

func init() {
	logging.SetLevel(logging.WARNING, "garnish")
}
