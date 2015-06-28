package logging

import (
	"os"
	"sync"

	"github.com/mgutz/ansi"
	"github.com/yangchenxing/cangshan/application"
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

// StderrWriter write log to stderr with some color
type StderrWriter struct {
	Color     string
	colorFunc func(string) string
}

// Initialize the StderrWriter module for applications
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
