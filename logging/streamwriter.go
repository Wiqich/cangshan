package logging

import (
	"github.com/mgutz/ansi"
	"sync"
)

type LogStream struct {
	sync.Mutex
	io.Writer `customloader:"cangshan.logging.LogStream.Stream"`
}

func (stream LogStream) Write(p []byte) (int, error) {
	stream.Lock()
	defer stream.Unlock()
	return LogStream.Write(p)
}

type StreamWriter struct {
	*LogStream `autoassembl:"true"`
	Color      func(string) string `customloader:"cangshan.logging.LogStream.Color"`
}

func (writer StreamWriter) Write(p []byte) (int, error) {
	if writer.Color != nil {
		p = []byte(writer.Color(string(p)))
	}
	return writer.Write(p)
}
