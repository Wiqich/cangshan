package tomlconfig

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/yangchenxing/cangshan/application"
	"io"
	"io/ioutil"
)

type includableConfig struct {
	Include []string
}

func LoadFile(path string) (map[string]interface{}, error) {
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
		if buf, err := ioutil.ReadFile(path); err != nil {
			return nil, fmt.Errorf("read config included file %s fail: %s", path, err.Error())
		} else {
			content.WriteByte('\n')
			content.Write(buf)
		}
	}
	config := make(map[string]interface{})
	if err := toml.Unmarshal(content.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("unmarshal full toml content fail: %s", err.Error())
	}
	return config, nil
}

func NewApplication(path string) (*Application, error) {
	if config, err := LoadFile(path); err != nil {
		return nil, err
	} else {
		return application.NewApplication(config)
	}
}
