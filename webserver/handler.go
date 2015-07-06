package webserver

// A Handler process webserver requests, modify request fields or generate response
type Handler interface {
	Handle(*Request)
}

type requestHandler struct{}

func (handler *requestHandler) Handle(request *Request) {
	request.handler.Handle(request)
}

var (
	requestHandlerSinglton = new(requestHandler)
)

type SimpleHandler func(*Request)

func (handler SimpleHandler) Handle(request *Request) {
	((func(*Request))(handler))(request)
}
