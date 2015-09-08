package iplocater_logging

import (
	"github.com/yangchenxing/cangshan/iplocater"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	iplocater.Debug = logging.Debug
}
