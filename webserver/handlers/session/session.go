package session

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/kv"
	"github.com/yangchenxing/cangshan/logging"
	"github.com/yangchenxing/cangshan/webserver"
)

const (
	DefaultSessionAttrKey   = "session"
	DefaultSessionIDAttrKey = "sessionid"
)

func init() {
	application.RegisterModulePrototype("WebServerSessionLoader", new(SessionLoader))
	application.RegisterModulePrototype("WebServerSessionSaver", new(SessionSaver))
}

type SessionLoader struct {
	SessionKey   string
	SessionIDKey string
	CookieName   string
	KV           kv.KV
	CookieMaxAge time.Duration
}

func (loader *SessionLoader) Initialize() error {
	if loader.SessionKey == "" {
		loader.SessionKey = DefaultSessionAttrKey
	}
	if loader.SessionIDKey == "" {
		loader.SessionIDKey = DefaultSessionIDAttrKey
	}
	return nil
}

func (loader SessionLoader) Handle(request *webserver.Request) {
	cookieName := loader.CookieName
	if cookieName == "" {
		cookieName := request.Host
		if pos := strings.Index(cookieName, ":"); pos > 0 {
			cookieName = cookieName[:pos]
		}
		cookieName += "_SESSION"
	}
	var session map[string]interface{}
	var sessionID string
	if cookie, err := request.Cookie(cookieName); err != nil || cookie == nil {
		sessionID = generateSessionID(request)
		cookie = new(http.Cookie)
		cookie.Name = cookieName
		cookie.Value = sessionID
		cookie.Path = "/"
		cookie.Domain = request.Host
		cookie.Expires = time.Now().Add(loader.CookieMaxAge)
		request.SetCookie(cookie)
		logging.Debug("Create session %s", sessionID)
		session = make(map[string]interface{})
	} else {
		sessionID = cookie.Value
		if data, err := loader.KV.Get(sessionID); err != nil {
			logging.Debug("Get session %s data fail: %s", sessionID, err.Error())
			session = make(map[string]interface{})
		} else if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&session); err != nil {
			logging.Warn("Decode session %s data fail: %s", sessionID, err.Error())
			session = make(map[string]interface{})
		}
	}
	request.Attr[loader.SessionKey] = session
	request.Attr[loader.SessionIDKey] = sessionID
}

func generateSessionID(request *webserver.Request) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, time.Now().Unix())
	binary.Write(buf, binary.LittleEndian, rand.Uint32())
	binary.Write(buf, binary.LittleEndian, rand.Uint32())
	return strings.ToUpper(hex.EncodeToString(buf.Bytes())[:32])
}

type SessionSaver struct {
	SessionKey    string
	SessionIDKey  string
	KV            kv.KV
	SessionMaxAge time.Duration
}

func (saver *SessionSaver) Initialize() error {
	if saver.SessionKey == "" {
		saver.SessionKey = DefaultSessionAttrKey
	}
	if saver.SessionIDKey == "" {
		saver.SessionIDKey = DefaultSessionIDAttrKey
	}
	return nil
}

func (saver SessionSaver) Handle(request *webserver.Request) {
	var data bytes.Buffer
	if sessionID, ok := request.Attr["sessionid"].(string); !ok {
		logging.Error("No session id")
	} else if session, ok := request.Attr["session"].(map[string]interface{}); !ok {
		logging.Error("No session %s", sessionID)
	} else if err := gob.NewEncoder(&data).Encode(session); err != nil {
		logging.Error("Encode session %s data fail: %s", sessionID, err.Error())
	} else if err := saver.KV.Set(sessionID, data.Bytes(), saver.SessionMaxAge); err != nil {
		logging.Error("Save session fail: %s", err.Error())
	}
}
