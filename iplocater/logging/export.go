package iplocater_logging

import (
	"github.com/yangchenxing/cangshan/iplocater"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	iplocater.Debug = logging.Debug
	iplocater.Info = logging.Info
	iplocater.Warn = logging.Warn
	iplocater.Error = logging.Error
	iplocater.Fatal = logging.Fatal
}
