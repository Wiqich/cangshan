package simplerest

import "github.com/yangchenxing/cangshan/webserver"

type FieldValidator interface {
	Validate(value interface{}, request *webserver.Request) error
}
