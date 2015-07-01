package logging

import (
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/email"
)

func init() {
	application.RegisterModulePrototype("EMailLogWriter", new(EMailWriter))
}

type EMailWriter struct {
	Client  email.EMail
	Subject string
}

func (w EMailWriter) Write(b []byte) (int, error) {
	if err := w.Client.SendText(w.Subject, string(b)); err != nil {
		return 0, err
	}
	return len(b), nil
}
