package logging

import (
	"github.com/pion/logging"
)

var loggerFactory = logging.NewDefaultLoggerFactory()

func NewLogger(scope string) logging.LeveledLogger {
	return loggerFactory.NewLogger(scope)
}
