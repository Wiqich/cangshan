package webserver

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("WebServer", new(WebServer))
	application.RegisterModulePrototype("WebServerLocation", new(Location))
}

// A Location match requests with specified path pattern, and provide handlers to handle the request
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

// Initialize the Location module for application
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
	loc.handlers[i] = requestHandlerSinglton
	i++
	for _, handler := range loc.PostProcess {
		loc.handlers[i] = handler
		i++
	}
	return nil
}

// Match check whether the url path match the location.
// Named groups of the matched path will be return as a map.
func (loc *Location) Match(request *Request) (bool, map[string]string) {
	if len(loc.methods) > 0 && !loc.methods[request.Method] {
		return false, nil
	}
	if loc.path == nil {
		if request.URL.Path == loc.Path {
			return true, nil
		}
		return false, nil
	} else if subexps := loc.path.FindStringSubmatch(request.URL.Path); subexps != nil {
		params := make(map[string]string)
		for i, name := range loc.path.SubexpNames() {
			if name != "" && subexps[i] != "" {
				params[name] = subexps[i]
			}
		}
		return true, params
	} else {
		return false, nil
	}
}

// A WebServer implements a web server module for application
type WebServer struct {
	*http.Server
	Name         string
	Locations    []*Location
	LogFormatter *logging.Formatter
}

func (server *WebServer) Initialize() error {
	if server.Server.Handler == nil {
		server.Server.Handler = server
	}
	return nil
}

func (server *WebServer) Run() error {
	logging.Info("Start web server %s", server.Name)
	return server.Server.ListenAndServe()
}

func (server *WebServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	server.serveHTTP(newRequest(request, response, server.LogFormatter))
}

func (server *WebServer) serveHTTP(request *Request) {
	defer request.buildResponse()
	for _, location := range server.Locations {
		if ok, params := location.Match(request); ok {
			request.Param = make(map[string]interface{})
			for key, value := range params {
				request.Param[key] = value
			}
			request.handler = location.Handler
			for _, handler := range location.handlers {
				handler.Handle(request)
				if request.stopped {
					return
				}
			}
			return
		}
	}
	request.Write(404, nil, "")
}
