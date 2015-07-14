package experiment

import (
	"github.com/yangchenxing/cangshan/logging"
	"math/rand"
	"reflect"
)

func init() {
	RegisterRouterType("Random", reflect.TypeOf(RandomRouter{}))
}

type RandomRouter struct {
	Choices []uint
}

func (router RandomRouter) SelectBranch(_ Features) (int, error) {
	value := uint(rand.Intn(1000) + 1)
	for i, percent := range router.Choices {
		if value <= percent {
			logging.Debug("random router choice %d", i)
			return i, nil
		}
		value -= percent
	}
	logging.Debug("random router choice -1")
	return -1, nil
}
