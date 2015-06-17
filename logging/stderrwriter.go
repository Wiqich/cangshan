package logging

import (
	"github.com/mgutz/ansi"
	"os"
	"sync"
)

type ColorFunc func(string) string

func (stream LogStream) Write(p []byte) (int, error) {
	stream.Lock()
	defer stream.Unlock()
	return LogStream.Write(p)
}

type StderrWriter struct {
	mutex sync.Mutex
	Color ColorFunc
}

func (writer StderrWriter) Write(p []byte) (int, error) {
	if writer.Color != nil {
		p = []byte(writer.Color(string(p)))
	}
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	return os.Stderr.Write(p)
}
