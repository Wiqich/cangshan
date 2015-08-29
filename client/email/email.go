package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("EMail", new(EMail))
}

type EMail struct {
	Server    string
	Username  string
	Password  string
	Sender    string
	Receivers []string
	auth      smtp.Auth
}

func (client *EMail) Initialize() error {
	client.auth = smtp.PlainAuth(client.Username, client.Username, client.Password,
		strings.Split(client.Server, ":")[0])
	return nil
}

func (client EMail) SendText(subject, content string, receivers ...string) error {
	var buffer bytes.Buffer
	receivers = append(receivers, client.Receivers...)
	fmt.Fprintf(&buffer, "From: %s\r\n", client.Sender)
	fmt.Fprintf(&buffer, "To: %s\r\n", strings.Join(receivers, ","))
	fmt.Fprintf(&buffer, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buffer, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buffer, "\r\n%s", content)
	err := smtp.SendMail(client.Server, client.auth, client.Sender, receivers, buffer.Bytes())
	if err != nil {
		return fmt.Errorf("Send mail fail: %s", err.Error())
	}
	return nil
}
