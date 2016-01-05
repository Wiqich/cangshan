package logging

import "fmt"

type Handler struct {
	Format    string
	Levels    []string
	Writers   []*writer
	formatter *formatter
}

func (handler *Handler) Initialize() error {
	handler.formatter = newFormatter(handler.Format)
	for i, writer := range handler.Writers {
		if err := writer.initialize(); err != nil {
			return fmt.Errorf("initialize writer[%d] fail: %s", i, err.Error())
		}
	}
	// fmt.Fprintf(os.Stderr, "initialize handler success: %v\n", *handler)
	return nil
}

func (handler *Handler) write(event map[string]string) {
	text := handler.formatter.format(event)
	for _, writer := range handler.Writers {
		writer.write(text)
	}
}
