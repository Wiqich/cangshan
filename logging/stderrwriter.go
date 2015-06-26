package logging

import (
	"github.com/chenxing/cangshan/application"
	"github.com/mgutz/ansi"
	"os"
	"sync"
)

func init() {
	application.RegisterModuleCreater("StderrLogWriter",
		func() interface{} {
			return new(StderrWriter)
		})
}

var (
	stderrMutex sync.Mutex
)

type StderrWriter struct {
	Color     string
	colorFunc func(string) string
}

func (writer *StderrWriter) Initialize() error {
	if writer.Color != "" {
		writer.colorFunc = ansi.ColorFunc(writer.Color)
	}
	return nil
}

func (writer StderrWriter) Write(p []byte) (int, error) {
	if writer.colorFunc != nil {
		p = []byte(writer.colorFunc(string(p)))
	}
	stderrMutex.Lock()
	defer stderrMutex.Unlock()
	return os.Stderr.Write(p)
}
