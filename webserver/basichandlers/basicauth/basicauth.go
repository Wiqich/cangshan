package basicauth

import (
	"http"
	"github.com/yangchenxing/cangshan/webserver"
	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModuleCreater("BasicAuth", func () interface{} { return new(BasicAuth) })
}

type BasicAuthenticator interface {
	Authenticate(string, string) bool
}

type BasicAuth struct {
	Authenticator *BasicAuthenticator
	Realm         string
}

func (auth *BasicAuth) Handle(request *webserver.Request) {
	username, password, ok := request.BasicAuth()
	if ok && auth.Authenticator.Authenticate(username, password) {
		return
	}
	request.ResponseHeader().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, auth.Realm))
	request.Write(http.StatusUnauthorized, nil, "")
}
