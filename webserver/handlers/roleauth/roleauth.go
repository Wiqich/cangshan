package roleauth

import (
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
	"github.com/yangchenxing/cangshan/webserver/handlers/session"
)

func init() {
	application.RegisterModulePrototype("WebServerRoleAuth", new(RoleAuth))
}

const (
	DefaultRoleKey = "role"
)

type RoleAuth struct {
	SessionKey string
	RoleKey    string
	Roles      []string
	roles      map[string]bool
}

func (auth *RoleAuth) Initialize() error {
	if auth.SessionKey == "" {
		auth.SessionKey = session.DefaultSessionAttrKey
	}
	if auth.RoleKey == "" {
		auth.RoleKey = DefaultRoleKey
	}
	auth.roles = make(map[string]bool)
	for _, role := range auth.Roles {
		auth.roles[role] = true
	}
	return nil
}

func (auth RoleAuth) Handle(request *webserver.Request) {
	if session, ok := request.Attr[auth.SessionKey].(map[string]interface{}); !ok || session == nil {
		request.Error("Missing session")
		request.WriteAndStop(500, nil, "")
		return
	} else if roles, ok := session[auth.RoleKey].([]string); !ok || roles == nil {
		request.Error("Missing role")
		request.WriteAndStop(500, nil, "")
		return
	} else {
		for _, role := range roles {
			if auth.roles[role] {
				return
			}
		}
		request.WriteAndStop(403, nil, "")
	}
}
