package filecontainer

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("StringSetFile", new(StringSetFile))
}

type StringSetFile struct {
	sync.Mutex
	Path          string
	CheckInterval time.Duration
	FailSleep     time.Duration
	data          map[string]bool
	timestamp     time.Time
}

func (file *StringSetFile) Initialize() error {
	if err := file.update(); err != nil {
		return err
	}
	go func() {
		for {
			time.Sleep(file.CheckInterval)
			if err := file.update(); err != nil {
				logging.Error("Update string set file \"%s\" fail: %s", file.Path, err.Error())
			}
			time.Sleep(file.FailSleep)
		}
	}()
	return nil
}

func (file *StringSetFile) Has(text string) bool {
	return file.data[text]
}

func (file *StringSetFile) update() error {
	file.Lock()
	defer file.Unlock()
	if info, err := os.Stat(file.Path); err != nil {
		return err
	} else if !info.ModTime().After(file.timestamp) {
		return nil
	} else {
		data := make(map[string]bool)
		f, err := os.Open(file.Path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if text := scanner.Text(); text != "" {
				data[text] = true
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		file.data = data
		file.timestamp = info.ModTime()
		logging.Info("String set file \"%s\" updated", file.Path)
	}
	return nil
}
