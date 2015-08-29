package filecontainer

import (
	"container/list"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("JSONFile", new(JSONFile))
}

type JSONFile struct {
	sync.Mutex
	Path            string
	CheckInterval   time.Duration
	FailSleep       time.Duration
	Type            reflect.Type
	Value           interface{}
	updateCallbacks *list.List
	timestamp       time.Time
}

func (file *JSONFile) Initialize() error {
	file.updateCallbacks = list.New()
	if err := file.update(); err != nil {
		return err
	}
	go func() {
		for {
			time.Sleep(file.CheckInterval)
			logging.Debug("check json file update: %s", file.Path)
			if err := file.update(); err != nil {
				logging.Error("Update JSON File \"%s\" fail: %s", file.Path, err.Error())
				time.Sleep(file.FailSleep)
			}
		}
	}()
	return nil
}

func (file *JSONFile) OnUpdate(callback func()) {
	if callback != nil {
		file.updateCallbacks.PushBack(callback)
	}
}

func (file *JSONFile) update() error {
	file.Lock()
	defer file.Unlock()
	if info, err := os.Stat(file.Path); err != nil {
		return err
	} else if !info.ModTime().After(file.timestamp) {
		return nil
	} else {
		value := reflect.New(file.Type).Interface()
		if content, err := ioutil.ReadFile(file.Path); err != nil {
			return err
		} else if err := json.Unmarshal(content, value); err != nil {
			return err
		}
		// if temp, err := json.MarshalIndent(value, "", "    "); err == nil {
		// 	logging.Debug("load json file: %s", temp)
		// }
		file.Value = value
		file.timestamp = info.ModTime()
		for callback := file.updateCallbacks.Front(); callback != nil; callback = callback.Next() {
			callback.Value.(func())()
		}
		logging.Info("JSON file \"%s\" updated", file.Path)
	}
	return nil
}
