package msgq

type Message struct {
	Type   string
	Header map[string]interface{}
	Body   []byte
}

type MessageQueue interface {
	Publish(message *Message) error
	Subscribe(category string) (<-chan *Message, <-chan error)
}
