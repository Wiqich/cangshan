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
