package webserver

// A Handler process webserver requests, modify request fields or generate response
type Handler interface {
	Handle(request *Request)
}

type MatchHandler interface {
	Handle(request *Request) (match bool)
}
