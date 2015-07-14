package webserver

import (
	"fmt"
	"regexp"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("WebServerLocation", new(Location))
}

type Location struct {
	Path        string
	Methods     []string
	PreProcess  []Handler
	Handler     Handler
	PostProcess []Handler
	path        *regexp.Regexp
	handlers    []Handler
	methods     map[string]bool
}

func (loc *Location) Initialize() error {
	var err error
	if loc.Path[0] == '^' || loc.Path[len(loc.Path)-1] == '$' {
		if loc.path, err = regexp.Compile(loc.Path); err != nil {
			return fmt.Errorf("Invalid path pattern: %s", err.Error())
		}
	}
	loc.methods = make(map[string]bool)
	for _, method := range loc.Methods {
		loc.methods[method] = true
	}
	loc.handlers = make([]Handler, len(loc.PreProcess)+1+len(loc.PostProcess))
	i := 0
	for _, handler := range loc.PreProcess {
		loc.handlers[i] = handler
		i++
	}
	loc.handlers[i] = loc.Handler
	i++
	for _, handler := range loc.PostProcess {
		loc.handlers[i] = handler
		i++
	}
	return nil
}

func (loc *Location) Handle(request *Request) (match bool) {
	if len(loc.methods) > 0 && !loc.methods[request.Method] {
		return false
	}
	if loc.path == nil {
		if request.URL.Path != loc.Path {
			return false
		}
	} else if subexps := loc.path.FindStringSubmatch(request.URL.Path); subexps == nil {
		return false
	} else {
		for i, name := range loc.path.SubexpNames() {
			if name != "" && subexps[i] != "" {
				request.Param[name] = subexps[i]
			}
		}
	}
	for _, handler := range loc.handlers {
		handler.Handle(request)
		if request.stopped {
			return true
		}
	}
	return true
}
