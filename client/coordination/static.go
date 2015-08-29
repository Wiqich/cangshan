package coordination

import (
	"errors"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("StaticCoordination", new(StaticCoordination))
}

type StaticCoordination struct {
	Servers map[string][]string
}

func (sc *StaticCoordination) Discover(dir string) ([]Node, error) {
	servers := sc.Servers[dir]
	nodes := make([]Node, len(servers))
	for i, server := range servers {
		temp := strings.SplitN(server, ",", 2)
		nodes[i].Key = temp[0]
		if len(temp) == 2 {
			nodes[i].Value = temp[1]
		}
	}
	return nodes, nil
}

func (sc *StaticCoordination) Register(dir, name, value string, ttl time.Duration) error {
	return errors.New("Not supported")
}

func (sc *StaticCoordination) Remove(dir, name string) error {
	return errors.New("Not supported")
}

func (sc *StaticCoordination) Wait(dir string) (*CoordinationEvent, error) {
	for {
		time.Sleep(time.Hour)
	}
	return nil, nil
}

func (sc *StaticCoordination) LongWait(dir string, receiveChan chan<- *CoordinationEvent, stopChan <-chan bool) error {
	<-stopChan
	return nil
}
