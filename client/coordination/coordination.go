package coordination

import "time"

type CoordinationEventType int

const (
	CreateNodeEvent CoordinationEventType = iota
	ModifyNodeEvent
	DeleteNodeEvent
)

type CoordinationEvent struct {
	Type      CoordinationEventType
	Key       string
	Value     string
	PrevValue string
}

type Node struct {
	Key   string
	Value string
}

type Coordination interface {
	Discover(dir string) (nodes []Node, err error)
	Register(dir, name, value string, ttl time.Duration) (err error)
	Remove(dir, name string) (err error)
	Wait(dir string) (event *CoordinationEvent, err error)
	LongWait(dir string, receiveChan chan<- *CoordinationEvent, stopChan <-chan bool) (err error)
}
