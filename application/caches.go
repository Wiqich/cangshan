package application

import (
	"reflect"
	"sync"
)

// A field represents a single field found in a struct.
// copy from BurntSushi/toml
type field struct {
	index []int
	typ   reflect.Type
}

func typeFields(typ reflect.Type, baseIndex []int) map[string]*field {
	depth := len(baseIndex)
	fields := make(map[string]*field)
	for i := 0; i < typ.NumField(); i += 1 {
		index := make([]int, depth+1)
		copy(index, baseIndex)
		index[depth] = i
		f := typ.Field(i)
		if f.Anonymous {
			var fs map[string]*field
			if f.Type.Kind() == reflect.Ptr {
				fs = typeFields(f.Type.Elem(), index)
			} else {
				fs = typeFields(f.Type, index)
			}
			for key, value := range fs {
				fields[key] = value
			}
		} else if f.Name[0] >= 'A' && f.Name[0] <= 'Z' {
			fields[f.Name] = &field{index, f.Type}
		}
	}
	return fields
}

var fieldCache struct {
	sync.RWMutex
	cache map[reflect.Type]map[string]*field
}

func cachedTypeFields(typ reflect.Type) map[string]*field {
	fieldCache.RLock()
	fields := fieldCache.cache[typ]
	fieldCache.RUnlock()
	if fields != nil {
		return fields
	}

	fields = typeFields(typ, []int{})

	fieldCache.Lock()
	if fieldCache.cache == nil {
		fieldCache.cache = make(map[reflect.Type]map[string]*field)
	}
	fieldCache.cache[typ] = fields
	fieldCache.Unlock()
	return fields
}
