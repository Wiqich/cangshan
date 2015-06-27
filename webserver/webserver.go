package webserver

import (
	"bitbucket.org/yangchenxing/cangshan/logging"
	"fmt"
	"net/http"
	"regexp"
)

type Location struct {
	Path        string
	PreProcess  []Handler
	Handler     Handler
	PostProcess []Handler
	path        *regexp.Regexp
	handlers    []Handler
}

func (loc *Location) Initialize() error {
	if loc.Path[0] == '^' || loc.Path[len(loc.Path)-1] == '$' {
		if path, err = regexp.Compile(loc.Path); err != nil {
			return fmt.Errorf("Invalid path pattern: %s", err.Error())
		} else {
			loc.path = path
		}
	}
	loc.handlers = make([]Handler, len(loc.PreProcess)+1+len(loc.PostProcess))
	i := 0
	for _, handler := range loc.PreProcess {
		loc.handlers[i] = handler
		i += 1
	}
	loc.handlers[i] = requestHandlerSinglton
	i += 1
	for _, handler := range loc.PostProcess {
		loc.handlers[i] = handler
		i += 1
	}
	return nil
}

func (loc *Location) Match(path string) (bool, map[string]string) {
	if loc.path == nil {
		if path == loc.Path {
			return true, nil
		} else {
			return false, nil
		}
	} else if subexps := loc.path.FindStringSubmatch(path); subexps != nil {
		params := make(map[string]string)
		for i, name := range loc.path.SubexpNames() {
			if name != "" {
				params[name] = subexps[i]
			}
		}
		return true, params
	} else {
		return false, nil
	}
}

type WebServer struct {
	*http.Server
	Locations    []Location
	LogFormatter *logging.Formatter
}

func (server *WebServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	server.Serve(newRequest(request, response, server.LogFormatter))
}

func (server *WebServer) serveHTTP(request *Request) {
	defer request.buildResponse()
	for _, location := range server.Locations {
		if ok, params := location.Match(request.URL.Path); ok {
			request.Param = make(map[string]interface{})
			for key, value := range params {
				request.Param[key] = value
			}
		}
		request.handler = location.Handler
		for _, handler := range location.handlers {
			handler.Handle(request)
			if request.stopped {
				return
			}
		}
	}
}
