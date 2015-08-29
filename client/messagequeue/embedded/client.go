package embedmsgq

import (
	"sync"

	"github.com/yangchenxing/cangshan/client/messagequeue"
)

var (
	chans = make(map[string]map[*Client]struct{})
	mutex sync.Mutex
	nop   struct{}
)

type Client struct {
	msgChan chan *msgq.Message
	errChan chan error
}

func (client *Client) Initialize() error {
	client.msgChan = make(chan *msgq.Message)
	client.errChan = make(chan error)
	return nil
}

func (client *Client) Public(message *msgq.Message) error {
	for client, _ := range chans[message.Type] {
		select {
		case client.msgChan <- message:
			return nil
		default:
			go func() { client.msgChan <- message }()
		}
	}
	return nil
}

func (client *Client) Subscribe(category string) (<-chan *msgq.Message, <-chan error) {
	mutex.Lock()
	defer mutex.Unlock()
	cateChans, found := chans[category]
	if !found {
		cateChans = make(map[*Client]struct{})
		chans[category] = cateChans
	}
	cateChans[client] = nop
	return client.msgChan, client.errChan
}
