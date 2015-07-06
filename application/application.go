package application

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type moduleError struct {
	name string
	err  error
}

var (
	builtinModules = make(map[string]interface{})
)

type Application struct {
	modules         map[string]interface{}
	waitings        map[string][]*assembler
	mutex           sync.Mutex
	assembleTimeout time.Duration
}

type assembler struct {
	*Application
	name string
	ch   chan interface{}
}

func RegisterBuiltinModule(name string, module interface{}) {
	builtinModules[name] = module
}

func NewApplication(config map[string]interface{}, timeout time.Duration) (*Application, error) {
	app := &Application{
		modules:         make(map[string]interface{}),
		waitings:        make(map[string][]*assembler),
		assembleTimeout: timeout,
	}
	errChan := make(chan *moduleError)
	count := 0
	for moduleType, config := range config {
		if moduleType == "alias" {
			var items []string
			asm := &assembler{
				Application: app,
				name:        "alias",
				ch:          make(chan interface{}),
			}
			if err := asm.unify(config, rvalue(&items)); err != nil {
				return nil, fmt.Errorf("invalid alias: %s", err.Error())
			}
			count += 1
			go func() {
				errChan <- asm.alias(items)
			}()
			continue
		}
		if nonModules[moduleType] {
			continue
		}
		if moduleCreater := moduleCreaters[moduleType]; moduleCreater == nil {
			return nil, fmt.Errorf("Unknown module type: %s", moduleType)
		} else if config, ok := config.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("Invalid module category %s config: not map[string]interface{}", moduleType)
		} else {
			for name, config := range config {
				name = moduleType + "." + name
				asm := &assembler{
					Application: app,
					name:        name,
					ch:          make(chan interface{}),
				}
				go asm.loadModule(name, config, moduleCreater.Create(), errChan)
				count += 1
			}
		}
	}
	for count > 0 {
		if err := <-errChan; err != nil {
			app.mutex.Lock()
			defer app.mutex.Unlock()
			waitChan := app.searchWaitingChan(err.name, map[string]bool{})
			return nil, fmt.Errorf("Load module %s fail: %s, waiting chan: %s", err.name,
				err.err.Error(), strings.Join(waitChan, " -> "))
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

func (asm *assembler) loadModule(name string, config interface{}, module interface{}, errChan chan<- *moduleError) {
	if err := asm.unify(config, rvalue(module)); err != nil {
		errChan <- &moduleError{name, err}
		return
	}
	if initializable, ok := module.(Initializable); ok {
		if err := initializable.Initialize(); err != nil {
			errChan <- &moduleError{name, fmt.Errorf("Initialize module %s fail: %s", name, err.Error())}
		}
	}
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	asm.modules[name] = module
	for _, waiting := range asm.waitings[name] {
		waiting.ch <- module
	}
	delete(asm.waitings, name)
	errChan <- nil
}

func (asm *assembler) getModuleOrWait(name string) (interface{}, <-chan interface{}) {
	if module, found := builtinModules[name]; found {
		return module, nil
	}
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	if module, found := asm.modules[name]; found {
		return module, nil
	}
	waitings, found := asm.waitings[name]
	if !found {
		waitings = make([]*assembler, 0, 1)
	}
	asm.waitings[name] = append(waitings, asm)
	return nil, asm.ch
}

func (asm *assembler) getModule(name string) (interface{}, error) {
	m, c := asm.getModuleOrWait(name)
	if c == nil {
		return m, nil
	}
	select {
	case m = <-c:
		break
	case <-time.After(asm.assembleTimeout):
		return nil, fmt.Errorf("wait module %s timeout", name)
	}
	return m, nil
}

func (asm *assembler) alias(items []string) *moduleError {
	for _, item := range items {
		fields := strings.Split(item, "->")
		if len(fields) != 2 {
			continue
		}
		name := strings.TrimSpace(fields[0])
		alias := strings.TrimSpace(fields[1])
		module, err := asm.getModule(name)
		if err != nil {
			return &moduleError{
				name: "alias",
				err:  fmt.Errorf("alias module %s to %s fail: %s", name, alias, err.Error()),
			}
		}
		asm.mutex.Lock()
		Debug("alias %s to %s", name, alias)
		asm.modules[alias] = module
		for _, waiting := range asm.waitings[alias] {
			waiting.ch <- module
		}
		asm.mutex.Unlock()
	}
	return nil
}

func (app *Application) searchWaitingChan(name string, searched map[string]bool) []string {
	found := "not found"
	if _, f := app.modules[name]; f {
		found = "found"
	}
	result := []string{fmt.Sprintf("%s(%s)", name, found)}
	searched[name] = true
	for depName, waitings := range app.waitings {
		for _, waitModule := range waitings {
			if waitModule.name == name {
				if searched[depName] {
					return append(result, depName+"(CYCLE!)")
				}
				return append(result, app.searchWaitingChan(depName, searched)...)
			}
		}
	}
	return result
}
