package webserver

type HookHandler interface {
	Handle(*Request) error
}

type RequestHandler struct {
}
