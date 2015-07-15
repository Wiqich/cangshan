package etcdcoordination

import (
	"errors"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/coordination"
)

func init() {
	application.RegisterModulePrototype("EtcdCoordination", new(EtcdCoordination))
}

type EtcdCoordination struct {
	Machines []string
	client   *etcd.Client
}

func (ec *EtcdCoordination) Initialize() error {
	if len(ec.Machines) == 0 {
		return errors.New("Missing Machines config")
	}
	ec.client = etcd.NewClient(ec.Machines)
	return nil
}

func (ec *EtcdCoordination) Discover(dir string) ([]coordination.Node, error) {
	response, err := ec.client.Get(dir, false, true)
	if err != nil {
		return nil, err
	}
	if response.Node == nil {
		return nil, errors.New("no Node in response")
	}
	nodes := make([]coordination.Node, len(response.Node.Nodes))
	for i, node := range response.Node.Nodes {
		nodes[i].Key = node.Key
		nodes[i].Value = node.Value
	}
	return nodes, nil
}

func (ec *EtcdCoordination) Register(dir, name, value string, ttl time.Duration) error {
	_, err := ec.client.Set(dir+"/"+name, value, uint64(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (ec *EtcdCoordination) Wait(dir string) (*coordination.CoordinationEvent, error) {
	response, err := ec.client.Watch(dir, 0, true, nil, nil)
	if err != nil {
		return nil, err
	}
	return ec.newEvent(response), nil
}

func (ec *EtcdCoordination) LongWait(dir string, receiveChan chan<- *coordination.CoordinationEvent, stopChan <-chan bool) error {
	rCh := make(chan *etcd.Response, 1)
	sCh := make(chan bool, 1)
	eCh := make(chan error, 1)
	go func() {
		_, err := ec.client.Watch(dir, 0, true, rCh, sCh)
		eCh <- err
	}()
	for {
		select {
		case response := <-rCh:
			receiveChan <- ec.newEvent(response)
		case stop := <-sCh:
			sCh <- stop
			return nil
		case err := <-eCh:
			return err
		}
	}
	return nil
}

func (ec *EtcdCoordination) newEvent(response *etcd.Response) *coordination.CoordinationEvent {
	switch response.Action {
	case "set":
		if response.PrevNode == nil {
			return &coordination.CoordinationEvent{
				Type:  coordination.CreateNodeEvent,
				Key:   response.Node.Key,
				Value: response.Node.Value,
			}
		} else {
			return &coordination.CoordinationEvent{
				Type:      coordination.ModifyNodeEvent,
				Key:       response.Node.Key,
				Value:     response.Node.Value,
				PrevValue: response.PrevNode.Value,
			}
		}
	case "delete":
		return &coordination.CoordinationEvent{
			Type:      coordination.DeleteNodeEvent,
			Key:       response.Node.Key,
			PrevValue: response.Node.Value,
		}
	}
	return nil
}
