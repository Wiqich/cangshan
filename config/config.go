package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"path/filepath"
)

var (
	configFile   = flag.String("config", "conf/config.toml", "path to config file")
	configFormat = flag.String("configFormat", "", "config file format, unset for auto detect")
)

func LoadGlobalConfig() error {
	if content, err := ioutil.ReadFile(*configFile); err != nil {
		return fmt.Errorf("read config file fail: %s", err.Error())
	} else {
		format := *configFormat
		if format == "" {
			format = filepath.Ext(filepath.Base(path))
		}
		conf := make(map[string]interface{})
		switch format {
		case "toml":
			if err := toml.Unmarshal(content, &conf); err != nil {
				return fmt.Errorf("unmarshal toml config file fail: %s", err.Error())
			}
		case "json":
			if err := json.Unmarshal(content, &conf); err != nil {
				return fmt.Errorf("unmarshal json config file fail: %s", err.Error())
			}
		}
	}

}
