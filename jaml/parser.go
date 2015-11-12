package jaml

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
)

func ParseFile(path string, debug func(string)) (*Value, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(content, debug, filepath.Dir(path))
}

func ParseBytes(content []byte, debug func(string), dir string) (*Value, error) {
	return ParseString(string(content), debug, dir)
}

func ParseString(content string, debug func(string), dir string) (*Value, error) {
	var debugs chan string
	if debug != nil {
		debugs = make(chan string)
		defer close(debugs)
		go func() {
			for log := range debugs {
				if log != "" {
					debug(log)
				}
			}
		}()
	}
	l := newLexer(content, debugs)
	if ret := yyParse(l); ret != 0 {
		if l.err != nil {
			return nil, l.err
		}
		return nil, fmt.Errorf("yacc fail: code=%d", ret)
	}
	root := <-l.root
	if err := root.DoImport(dir, debug); err != nil {
		return nil, err
	}
	return root, nil
}

func ParseStream(r io.Reader, debug func(string), dir string) (*Value, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseBytes(content, debug, dir)
}

func EnableYaccErrorVerbose() {
	yyErrorVerbose = true
}
