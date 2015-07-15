package application

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yangchenxing/cangshan/structs"
)

var (
	builtinModules = make(map[string]interface{})
	ErrDeadlock    = errors.New("application module depencency dead lock")
)

type Application struct {
	sync.Mutex
	modules  map[string]interface{}
	consts   map[string]interface{}
	run      []string
	waitings map[string][]namedChan
	events   chan *assemblerEvent
}

func RegisterBuiltinModule(name string, module interface{}) {
	builtinModules[name] = module
}

func NewApplication(config map[string]interface{}) (*Application, error) {
	app := &Application{
		modules:  make(map[string]interface{}),
		waitings: make(map[string][]namedChan),
		events:   make(chan *assemblerEvent, 1),
	}
	unfinished := 0
	for moduleType, config := range config {
		switch moduleType {
		case "alias":
			unfinished++
			go app.newAssembler("alias").alias(config)
			continue
		case "const":
			unfinished++
			go app.newAssembler("const").setConst(config)
			continue
		case "run":
			if err := structs.Unmarshal(config, &app.run); err != nil {
				return app, fmt.Errorf("config run fail: %s", err.Error())
			}
			continue
		}
		if moduleCreater := moduleCreaters[moduleType]; moduleCreater == nil {
			return app, fmt.Errorf("Unknown module type: %s", moduleType)
		} else if config, ok := config.(map[string]interface{}); !ok {
			return app, fmt.Errorf("Invalid module category %s config: not map[string]interface{}", moduleType)
		} else {
			for name, config := range config {
				name = moduleType + "." + name
				unfinished++
				go app.newAssembler(name).loadModule(config, moduleCreater.Create())
			}
		}
	}
	locked := 0
	for unfinished > 0 {
		if locked == unfinished {
			return app, ErrDeadlock
		}
		ev := <-app.events
		switch ev.typ {
		case waitEvent:
			locked++
		case receiveEvent:
			locked--
		case doneEvent:
			unfinished--
		}
		if ev.err != nil {
			return app, fmt.Errorf("create application fail during load module %s: %s",
				ev.name, ev.err.Error())
		}
	}
	return app, nil
}

func (app *Application) Run() error {
	errChan := make(chan error)
	count := 0
	for _, name := range app.run {
		if module := app.modules[name]; module == nil {
			return fmt.Errorf("Missing run module: %s", name)
		} else if run, ok := app.modules[name].(Runable); !ok {
			return fmt.Errorf("Module %s is not runable", name)
		} else {
			go func() {
				errChan <- run.Run()
			}()
			count += 1
		}
	}
	for i := 0; i < count; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) newAssembler(name string) *assembler {
	return &assembler{
		Application: app,
		name:        name,
	}
}

func (app *Application) DumpWatingSequences() [][]string {
	result := make([][]string, 0, 1)
	depends := make(map[string]string)
	for depName, waitings := range app.waitings {
		for _, waiting := range waitings {
			depends[waiting.name] = depName
		}
	}
	searched := make(map[string]bool)
	for name, _ := range depends {
		if searched[name] {
			continue
		}
		seq := []string{name}
		for dep, found := depends[name]; found; dep, found = depends[name] {
			searched[name] = true
			seq = append(seq, dep)
			if searched[dep] {
				seq = append(seq, dep)
				break
			}
			name = dep
		}
		result = append(result, seq)
	}

	return result
}
