package webserver

import (
	"errors"
)

type Handler interface {
	Handle(*Request)
}

type requestHandler struct{}

func (handler *requestHandler) Handler(request *Request) {
	request.handler(request)
}

var (
	requestHandlerSinglton = new(requestHandler)
)
