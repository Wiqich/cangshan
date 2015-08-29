package eventhandler

import (
	"os"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/coordination"
	"github.com/yangchenxing/cangshan/supervisor"
)

func init() {
	application.RegisterModulePrototype("SupervisorEventCoordinationHandler", new(CoordinationHandler))
}

type CoordinationHandler struct {
	Coordination coordination.Coordination
	Host         string
}

func (handler *CoordinationHandler) Initialize() error {
	if handler.Host == "" {
		handler.Host, _ = os.Hostname()
	}
	return nil
}

func (handler CoordinationHandler) Hande(event supervisor.Event) error {
	if handler.Host == "" {
		return nil
	}
	switch event.State() {
	case "RUNNING":
		if err := handler.Coordination.Register(event["processname"], handler.Host, "", 0); err != nil {
			return err
		}
	case "EXITED":
		fallthrough
	case "STOPPED":
		if err := handler.Coordination.Remove(event["processname"], handler.Host); err != nil {
			return err
		}
	}
	return nil
}
