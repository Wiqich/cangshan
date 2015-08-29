package logging

import (
	"bytes"
	"container/list"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/email"
)

const (
	defaultEMailWriterEventChannelCapacity = 64
	defaultEMailWriterMessageDelay         = time.Second * 5
	emailWriterLogPrefix                   = "[EMailWriter log]"
)

var (
	emailWriterLogPrefixBytes = []byte(emailWriterLogPrefix)
)

func init() {
	application.RegisterModulePrototype("EMailLogWriter", new(EMailWriter))
}

type EMailWriter struct {
	Client       *email.EMail
	Subject      string
	ChanCapacity int
	MessageDelay time.Duration
	eventChan    chan []byte
}

func (writer *EMailWriter) Initialize() error {
	if writer.ChanCapacity == 0 {
		writer.ChanCapacity = defaultEMailWriterEventChannelCapacity
	}
	if writer.MessageDelay == 0 {
		writer.MessageDelay = defaultEMailWriterMessageDelay
	}
	writer.eventChan = make(chan []byte, defaultEMailWriterEventChannelCapacity)
	go writer.sendMessage()
	return nil
}

func (writer EMailWriter) Write(b []byte) (int, error) {
	if bytes.Index(b, emailWriterLogPrefixBytes) == -1 {
		writer.eventChan <- b
	}
	return len(b), nil
}

func (writer *EMailWriter) sendMessage() {
	for {
		message := <-writer.eventChan
		Debug("%s recieve email message: %s", emailWriterLogPrefix, strings.TrimSpace(string(message)))
		messages := list.New()
		messages.PushBack(message)
		ticker := time.NewTicker(writer.MessageDelay)
		delay := true
		for delay {
			select {
			case message := <-writer.eventChan:
				Debug("%s recieve email message: %s", emailWriterLogPrefix, strings.TrimSpace(string(message)))
				messages.PushBack(message)
			case <-ticker.C:
				ticker.Stop()
				delay = false
				break
			}
		}
		// Debug("%s format email log", emailWriterLogPrefix)
		var body bytes.Buffer
		for message := messages.Front(); message != nil; message = message.Next() {
			body.Write(message.Value.([]byte))
		}
		// Debug("%s format email log done: %d bytes", emailWriterLogPrefix, body.Len())
		if err := writer.Client.SendText(writer.Subject, body.String()); err != nil {
			Error("%s send email fail: %s", emailWriterLogPrefix, err.Error())
		} else {
			Debug("%s send email success: %d bytes", emailWriterLogPrefix, body.Len())
		}
	}
}
