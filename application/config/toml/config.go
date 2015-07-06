package tomlapp

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/yangchenxing/cangshan/application"
	"io/ioutil"
	"path/filepath"
	"time"
)

type includableConfig struct {
	Include []string `toml:"include"`
}

func loadContent(path string) ([]byte, error) {
	dir := filepath.Dir(path)
	main, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %s fail: %s", path, err.Error())
	}
	var ic includableConfig
	if err = toml.Unmarshal(main, &ic); err != nil {
		return nil, fmt.Errorf("unmarshal toml file %s fail: %s", path, err.Error())
	}
	content := bytes.NewBuffer(main)
	for _, path := range ic.Include {
		path = filepath.Clean(path)
		if path[0] != '/' {
			path = filepath.Join(dir, path)
		}
		if subContent, err := loadContent(path); err != nil {
			return nil, err
		} else {
			content.WriteByte('\n')
			content.Write(subContent)
		}
	}
	return content.Bytes(), nil
}

func NewApplication(path string, timeout time.Duration) (*application.Application, error) {
	application.RegisterNonModule("include")
	content, err := loadContent(path)
	if err != nil {
		return nil, err
	}
	config := make(map[string]interface{})
	if err := toml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("unmarshal full toml content fail: %s", err.Error())
	}
	return application.NewApplication(config, timeout)
}
