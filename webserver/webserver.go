package webserver

import (
	"net/http"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("WebServer", new(WebServer))
}

// A WebServer implements a web server module for application
type WebServer struct {
	*http.Server
	Name         string
	Handlers     []MatchHandler
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
	for _, handler := range server.Handlers {
		if handler.Handle(request) {
			return
		}
	}
	request.Write(404, nil, "")
}
