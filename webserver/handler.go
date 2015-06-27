package webserver

import (
	"errors"
)

var (
	ErrRewrite = errors.New("rewrite")
	ErrStop    = errors.New("stop")
)

type HookHandler interface {
	HandleHook(*Request) error
}

type RequestHandler interface {
	HandleRequest(*Request)
}
