package basicauth

import (
	"fmt"
	"net/http"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
)

func init() {
	application.RegisterModulePrototype("WebServerBasicAuthHandler", new(BasicAuth))
	application.RegisterModulePrototype("WebServerSimpleBasicAuthenticator", new(SimpleBasicAuthenticator))
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

type User struct {
	Username string
	Password string
}

type SimpleBasicAuthenticator struct {
	Users []User
	users map[string]string
}

func (auth *SimpleBasicAuthenticator) Initialize() error {
	auth.users = make(map[string]string)
	for _, u := range auth.Users {
		auth.users[u.Username] = u.Password
	}
	return nil
}

func (auth *SimpleBasicAuthenticator) Authenticate(username, password string) bool {
	return auth.users[username] == password
}
