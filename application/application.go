package application

import (
	"fmt"
	"sync"
)

type moduleError struct {
	name string
	err  error
}

type Application struct {
	modules  map[string]interface{}
	waitings map[string][]chan<- interface{}
	mutex    sync.Mutex
}

func NewApplication(config map[string]interface{}) (*Application, error) {
	app := &Application{
		modules:  make(map[string]interface{}),
		waitings: make(map[string][]chan<- interface{}),
	}
	errChan := make(chan *moduleError)
	count := 0
	for moduleType, config := range config {
		if moduleCreater := moduleCreaters[moduleType]; moduleCreater == nil {
			return nil, fmt.Errorf("Unknown module type: %s", moduleType)
		} else if config, ok := config.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("Invalid module category %s config: not map[string]interface{}", moduleType)
		} else {
			for name, config := range config {
				go app.loadModule(moduleType+"."+name, config, moduleCreater(), errChan)
				count += 1
			}
		}
	}
	for count > 0 {
		if err := <-errChan; err != nil {
			return nil, fmt.Errorf("Load module %s fail: %s", err.name, err.err.Error())
		}
		count -= 1
	}
	return app, nil
}

func (app *Application) Run() error {
	runables := make([]Runable, 0, 1)
	for _, module := range app.modules {
		if runable, ok := module.(Runable); ok {
			runables = append(runables, runable)
		}
	}
	errChan := make(chan error, len(runables))
	for _, runable := range runables {
		go func() {
			errChan <- runable.Run()
		}()
	}
	for range runables {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) loadModule(name string, config interface{}, module interface{}, errChan chan<- *moduleError) {
	if err := app.unify(config, rvalue(module)); err != nil {
		errChan <- &moduleError{name, err}
		return
	}
	if initializable, ok := module.(Initializable); ok {
		if err := initializable.Initialize(); err != nil {
			errChan <- &moduleError{name, fmt.Errorf("Initialize module %s fail: %s", name, err.Error())}
		}
	}
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.modules[name] = module
	for _, waiting := range app.waitings[name] {
		waiting <- module
	}
	errChan <- nil
}

func (app *Application) getModule(name string) <-chan interface{} {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	ch := make(chan interface{}, 1)
	if module, found := app.modules[name]; found {
		ch <- module
		return ch
	}
	waitings, found := app.waitings[name]
	if !found {
		waitings = make([]chan<- interface{}, 0, 1)
	}
	app.waitings[name] = append(waitings, ch)
	return ch
}
