package logging

import (
	"errors"
	"github.com/chenxing/cangshan/application"
	"io"
)

func init() {
	application.RegisterModuleCreater("LogHandler",
		func() interface{} {
			return new(Handler)
		})
}

type Handler struct {
	Formatter *Formatter
	Levels    []string
	Writer    io.Writer
}

func (handler *Handler) Initialize() error {
	if handler.Formatter == nil {
		return errors.New("missing formatter")
	} else if len(handler.Levels) == 0 {
		return errors.New("missing levels")
	} else if handler.Writer == nil {
		return errors.New("missing writer")
	}
	return nil
}

func (handler *Handler) write(e *event) {
	handler.Writer.Write([]byte(handler.Formatter.format(e)))
}

func createDefaultHandler() *Handler {
	formatter := &Formatter{
		Format: "%level [%time][%filename:%line][%funcname] %message",
	}
	formatter.Initialize()
	return &Handler{
		Formatter: formatter,
		Levels:    []string{"debug", "info", "warn", "error", "fatal"},
		Writer:    &StderrWriter{},
	}
}
