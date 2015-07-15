package tomlapp

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/structs"
)

func NewApplication(path string) (*application.Application, error) {
	conf, err := loadContent(path, make(map[string]bool))
	if err != nil {
		return nil, err
	}
	return application.NewApplication(conf)
}

func loadContent(path string, loaded map[string]bool) (map[string]interface{}, error) {
	if loaded[path] {
		return nil, nil
	}
	loaded[path] = true
	dir := filepath.Dir(path)
	var conf map[string]interface{}
	if content, err := ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("Read config fail %s fail: %s", path, err.Error())
	} else if err := toml.Unmarshal(content, &conf); err != nil {
		return nil, fmt.Errorf("Unmarshal TOML file %s fail: %s", path, err.Error())
	} else {
		if include := conf["include"]; include != nil {
			delete(conf, "include")
			var includes []string
			if err := structs.Unmarshal(include, &includes); err != nil {
				return nil, fmt.Errorf("Invalid include in config file %s: %s", path, err.Error())
			}
			for _, includeFile := range includes {
				includeFile = filepath.Join(dir, includeFile)
				if includeConf, err := loadContent(includeFile, loaded); err != nil {
					return nil, fmt.Errorf("Load include file fail: %s", err.Error())
				} else if err := merge(conf, includeConf); err != nil {
					return nil, fmt.Errorf("Merge config file %s and %s fail: %s", path, includeFile, err.Error())
				}
			}
		}
	}
	return conf, nil
}

func merge(main, include map[string]interface{}) error {
	if include == nil {
		return nil
	}
	for key, right := range include {
		if left, found := main[key]; !found {
			main[key] = right
		} else {
			leftValue := reflect.ValueOf(left)
			rightValue := reflect.ValueOf(right)
			leftType := leftValue.Type()
			rightType := rightValue.Type()
			switch leftValue.Kind() {
			case reflect.Slice:
				if rightValue.Kind() != reflect.Slice || leftType.Elem() != rightType.Elem() {
					return fmt.Errorf("cannot merge %t to %t", right, left)
				}
				leftValue.Set(reflect.AppendSlice(leftValue, rightValue))
			case reflect.Map:
				if rightValue.Kind() != reflect.Map || leftType.Key() != rightType.Key() || leftType.Elem() != rightType.Elem() {
					return fmt.Errorf("cannot merge %t to %t", right, left)
				}
				for _, keyValue := range rightValue.MapKeys() {
					leftValue.SetMapIndex(keyValue, rightValue.MapIndex(keyValue))
				}
			default:
				return fmt.Errorf("Not mergable key \"%s\" of type \"%s\"", key, leftValue.Type())
			}
		}
	}
	return nil
}
