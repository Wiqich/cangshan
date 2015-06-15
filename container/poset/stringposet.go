package container

import (
	"errors"
	"github.com/chenxing/cangshan/container/stringset"
)

type StringPoset map[string]stringset.StringSet

func NewStringPoset() StringPoset {
	return make(map[string]stringset.StringSet)
}

func (poset StringPoset) Add(pre, post string) error {
	if pre == "" || post == "" {
		return errors.New("empty value")
	}
	if _, found := poset[post]; !found {
		poset[post] = stringset.New()
	}
	poset[post].Add(pre)
	return nil
}

func (poset StringPoset) Pop() string {
	result := poset.pop()
	delete(poset, result)
	for _, value := range poset {
		delete(value, result)
	}
	return result
}

func (poset StringPoset) Len() int {
	return len(poset)
}

func (poset StringPoset) pop() string {
	for key, value := range poset {
		if value.Len() == 0 {
			return key
		}
	}
	return ""
}
