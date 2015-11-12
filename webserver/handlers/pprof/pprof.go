package pprof

import (
	"net/http"
	gopprof "net/http/pprof"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
)

func init() {
	application.RegisterBuiltinModule("WebServerPProfIndex", pprofWrapper(gopprof.Index))
	application.RegisterBuiltinModule("WebServerPProfCmdline", pprofWrapper(gopprof.Cmdline))
	application.RegisterBuiltinModule("WebServerPProfProfile", pprofWrapper(gopprof.Profile))
	application.RegisterBuiltinModule("WebServerPProfSymbol", pprofWrapper(gopprof.Symbol))
}

type pprofWrapper func(http.ResponseWriter, *http.Request)

func (wrapper pprofWrapper) Handle(request *webserver.Request) {
	wrapper(request.GetHttpResponseWriter(), request.GetHttpRequest())
	request.Done()
}
