package supervisor

import (
	"fmt"
	"io"
	"os"

	"github.com/yangchenxing/cangshan/logging"
)

type EventHandler interface {
	Handle(event Event) error
}

type EventListener struct {
	Handlers map[string][]EventHandler
	Reader   io.Reader
	Writer   io.Writer
}

func (listener EventListener) Run() error {
	if listener.Reader == nil {
		listener.Reader = os.Stdin
	}
	if listener.Writer == nil {
		listener.Writer = os.Stdout
	}
	for {
		fmt.Fprintf(listener.Writer, "READY\n")
		event, err := NewEvent(listener.Reader)
		if err != nil {
			return fmt.Errorf("New event fail: %s", err.Error())
		}
		for _, handler := range listener.Handlers[event.Name()] {
			if err := handler.Handle(event); err != nil {
				logging.Error("handle event %s fail: %s", event.Name(), err.Error())
			}
		}
		for _, handler := range listener.Handlers[event.AbstractType()] {
			if err := handler.Handle(event); err != nil {
				logging.Error("handle abstract event %s fail: %s", event.Name(), err.Error())
			}
		}
		fmt.Fprintf(listener.Writer, "RESULT 2\nOK")
	}
	return nil
}
