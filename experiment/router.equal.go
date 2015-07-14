package experiment

import (
	"fmt"
	"github.com/yangchenxing/cangshan/logging"
	"reflect"
)

func init() {
	RegisterRouterType("EqualInt", reflect.TypeOf(EqualIntRouter{}))
	RegisterRouterType("EqualString", reflect.TypeOf(EqualStringRouter{}))
}

type EqualIntRouter struct {
	FeatureNames []string
	Values       [][]int64
	values       []map[int64]bool
}

func (router *EqualIntRouter) Initialize() error {
	router.values = make([]map[int64]bool, len(router.Values))
	for i, values := range router.Values {
		router.values[i] = make(map[int64]bool)
		for _, value := range values {
			router.values[i][value] = true
		}
	}
	return nil
}

func (router EqualIntRouter) SelectBranch(features Features) (int, error) {
	for _, name := range router.FeatureNames {
		var value int64
		v := features.GetFeature(name)
		if v == nil {
			continue
		}
		switch v.(type) {
		case int:
			value = int64(v.(int))
		case int8:
			value = int64(v.(int8))
		case int16:
			value = int64(v.(int16))
		case int32:
			value = int64(v.(int32))
		case int64:
			value = v.(int64)
		case uint:
			value = int64(v.(uint))
		case uint8:
			value = int64(v.(uint8))
		case uint16:
			value = int64(v.(uint16))
		case uint32:
			value = int64(v.(uint32))
		case uint64:
			value = int64(v.(uint64))
		default:
			return -1, fmt.Errorf("feature %s is not integer", name)
		}
		for i, set := range router.values {
			if set[value] {
				logging.Debug("integer equal router choice %d by feature %s", i, name)
				return i, nil
			}
		}
	}
	logging.Debug("integer equal router choice -1")
	return -1, nil
}

type EqualStringRouter struct {
	FeatureNames []string
	Values       [][]string
	values       []map[string]bool
}

func (router *EqualStringRouter) Initialize() error {
	router.values = make([]map[string]bool, len(router.Values))
	for i, values := range router.Values {
		router.values[i] = make(map[string]bool)
		for _, value := range values {
			router.values[i][value] = true
		}
	}
	return nil
}

func (router EqualStringRouter) SelectBranch(features Features) (int, error) {
	for _, name := range router.FeatureNames {
		v := features.GetFeature(name)
		if v == nil {
			continue
		}
		value, ok := v.(string)
		if !ok {
			return -1, fmt.Errorf("feature %s is not string", name)
		}
		for i, set := range router.values {
			if set[value] {
				logging.Debug("string equal router choice %d by feature %s", i, name)
				return i, nil
			}
		}
	}
	logging.Debug("string equal router choice -1")
	return -1, nil
}
