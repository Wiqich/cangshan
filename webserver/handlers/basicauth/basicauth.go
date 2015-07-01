package basicauth

import (
	"fmt"
	"net/http"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
)

func init() {
	application.RegisterModuleCreater("WebServerBasicAuthHandler",
		func() interface{} {
			return new(BasicAuth)
		})
}

// A BasicAuthenticator can authenticate username with password.
type BasicAuthenticator interface {
	Authenticate(username, password string) bool
}

// A BasicAuth implement HTTP basic auth.
type BasicAuth struct {
	Authenticator BasicAuthenticator
	Realm         string
}

// Handle basic auth of the request, the request should stop with 401 status if the authentication
// fail.
func (auth *BasicAuth) Handle(request *webserver.Request) {
	username, password, ok := request.BasicAuth()
	if ok && auth.Authenticator.Authenticate(username, password) {
		return
	}
	request.ResponseHeader().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, auth.Realm))
	request.Write(http.StatusUnauthorized, nil, "")
}
