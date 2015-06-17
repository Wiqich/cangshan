package unmarshaler

import (
	"fmt"
)

type unmarshalError struct {
	err  error
	path string
}

func newUnmarshalError(format string, params ...interface{}) *unmarshalError {
	return &unmarshalError{
		err:  fmt.Errorf(format, params...),
		path: "",
	}
}

func (err *unmarshalError) wrap(path string) *unmarshalError {
	if err.path != "" {
		if err.path[0] == '[' {
			path += err.path
		} else {
			path += "." + err.path
		}
	}
	return &unmarshalError{
		err:  err.err,
		path: path,
	}
}
