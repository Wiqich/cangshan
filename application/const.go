package application

import (
	"errors"
	"fmt"

	"github.com/yangchenxing/cangshan/structs"
)

func (asm *assembler) setConst(config interface{}) {
	var consts []struct {
		Name  string
		Value interface{}
	}
	if err := structs.Unmarshal(config, &consts); err != nil {
		asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Unmarshal fail: %s", err.Error()))
		return
	}
	asm.Lock()
	defer asm.Unlock()
	for _, c := range consts {
		if c.Name == "" || c.Value == nil {
			asm.events <- asm.newEvent(doneEvent, errors.New("Missing \"Name\" or \"Value\""))
			return
		}
		asm.consts[c.Name] = c.Value
	}
	for _, waiting := range asm.waitings["const"] {
		asm.events <- asm.newEvent(receiveEvent, nil)
		waiting.ch <- nil
	}
	delete(asm.waitings, "const")
}
