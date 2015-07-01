package logging

import (
	"errors"
	"io"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("LogHandler", new(Handler))
}

// A Handler convert acceptable log event to string with Formatter and write to some io.Writer
type Handler struct {
	Formatter *Formatter
	Levels    []string
	Writer    io.Writer
}

// Initialize Handler module for application
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

func (handler *Handler) write(e *event, formatter *Formatter) {
	if formatter == nil {
		formatter = handler.Formatter
	}
	handler.Writer.Write([]byte(formatter.format(e)))
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
