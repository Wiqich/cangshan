package webserver

import "github.com/yangchenxing/cangshan/application"

func init() {
	application.RegisterModulePrototype("WebServerHandlerGroup", new(HandlerGroup))
}

type HandlerGroup struct {
	Handlers []MatchHandler
}

func (group *HandlerGroup) Handle(request *Request) (match bool) {
	for _, handler := range group.Handlers {
		if handler.Handle(request) {
			return true
		}
	}
	return false
}
