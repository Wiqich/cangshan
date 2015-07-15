package application

import (
	"errors"
	"fmt"

	"github.com/yangchenxing/cangshan/structs"
)

func (asm *assembler) alias(config interface{}) {
	var alias []struct {
		Name  string
		Alias string
	}
	if err := structs.Unmarshal(config, &alias); err != nil {
		asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Unmarshal fail: %s", err.Error()))
		return
	}
	for _, alias := range alias {
		if alias.Name == "" || alias.Alias == "" {
			asm.events <- asm.newEvent(doneEvent, errors.New("Missing \"Name\" or \"Alias\""))
			return
		}
		module := asm.getModule(alias.Name)
		asm.Lock()
		asm.modules[alias.Alias] = module
		for _, waiting := range asm.waitings[alias.Alias] {
			asm.events <- asm.newEvent(receiveEvent, nil)
			waiting.ch <- module
		}
		delete(asm.waitings, alias.Alias)
		asm.Unlock()
	}
	asm.events <- asm.newEvent(doneEvent, nil)
}
