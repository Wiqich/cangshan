package eventhandler

import (
	"github.com/yangchenxing/cangshan/client/email"
	"github.com/yangchenxing/cangshan/logging"
	"github.com/yangchenxing/cangshan/strings"
	"github.com/yangchenxing/cangshan/supervisor"
)

type EMailHandler struct {
	EMail     email.EMail
	Subject   string
	Content   string
	Receivers []string
	subject   *stringutil.MapFormatter
	content   *stringutil.MapFormatter
}

func (handler *EMailHandler) Initialize() error {
	handler.subject = stringutil.NewMapFormatter(handler.Subject)
	handler.content = stringutil.NewMapFormatter(handler.Content)
	return nil
}

func (handler EMailHandler) Handle(event supervisor.Event) error {
	ev := make(map[string]interface{})
	for key, value := range event {
		ev[key] = value
	}
	err := handler.EMail.SendText(
		handler.subject.Format(ev),
		handler.content.Format(ev),
		handler.Receivers...)
	if err != nil {
		logging.Error("send event email fail: %s", err.Error())
	} else {
		logging.Info("send event email success: %s", event)
	}
	return nil
}
