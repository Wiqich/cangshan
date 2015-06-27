package webserver

import (
	"bitbucket.org/chenxing/cangshan/logging"
	"fmt"
	"net/http"
	"regexp"
)

type Location struct {
	Path             string
	PreProcessPhase  []*HookHandler
	RequestHandler   *RequestHandler
	PostProcessPhase []*HookHandler
	path             *regexp.Regexp
}

func (loc *Location) Initialize() error {
	if loc.Path[0] == '^' || loc.Path[len(loc.Path)-1] == '$' {
		if path, err = regexp.Compile(loc.Path); err != nil {
			return fmt.Errorf("Invalid path pattern: %s", err.Error())
		} else {
			loc.path = path
		}
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
	RewriteLimit int
}

func (server *WebServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	server.Serve(newRequest(request, response, server.LogFormatter))
}

func (server *WebServer) serveHTTP(request *Request) {
	// check rewrite limitation
	if request.rewriteCount >= server.RewriteLimit {
		request.Error("too many rewrite")
		request.Write(500, nil, "")
		request.buildResponse()
		return
	}
	for _, location := range server.Locations {
		if ok, params := location.Match(request.URL.Path); ok {
			request.Param = make(map[string]interface{})
			for key, value := range params {
				request.Param[key] = value
			}
		}
		for _, hook := range location.PreProcessPhase {
			if err := hook.HandleHook(request); err == ErrRewrite {
				request.rewriteCount += 1
			}
		}
		location.RequestHandler.HandleRequest(request)
		for _, hook := range location.PostProcessPhase {

		}
		request.buildResponse()
		break
	}
}
