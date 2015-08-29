package structs

import (
	"encoding"
	"fmt"
	"reflect"
	"time"
)

type UnmarshalMapValueHock func(data interface{}, rv reflect.Value) (newData interface{}, stop bool, err error)

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

func Unmarshal(data interface{}, value interface{}) error {
	return unmarshaler{nil}.unmarshal(data, rvalue(value))
}

func UnmarshalWithHock(data interface{}, value interface{}, hock UnmarshalMapValueHock) error {
	return unmarshaler{hock}.unmarshal(data, rvalue(value))
}

func UnmarshalValue(data interface{}, rv reflect.Value) error {
	return unmarshaler{nil}.unmarshal(data, rv)
}

func UnmarshalValueWithHock(data interface{}, rv reflect.Value, hock UnmarshalMapValueHock) error {
	return unmarshaler{hock}.unmarshal(data, rv)
}

type unmarshaler struct {
	hock UnmarshalMapValueHock
}

func (u unmarshaler) unmarshal(data interface{}, rv reflect.Value) error {
	var stop bool
	var err error
	if data, stop, err = u.invokeHock(data, rv); err != nil {
		return err
	} else if stop {
		return nil
	}
	if rv.CanAddr() {
		if v, ok := rv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.unmarshalText(data, v)
		}
	}
	if rv.Type() == timeType {
		return u.unmarshalTime(data, rv)
	}
	if rv.Type() == durationType {
		return u.unmarshalDuration(data, rv)
	}

	k := rv.Kind()

	if k >= reflect.Int && k <= reflect.Uint64 {
		return u.unmarshalInt(data, rv)
	}
	switch k {
	case reflect.Ptr:
		elem := reflect.New(rv.Type().Elem())
		if err := u.unmarshal(data, reflect.Indirect(elem)); err != nil {
			return err
		}
		rv.Set(elem)
		return nil
	case reflect.Struct:
		return u.unmarshalStruct(data, rv)
	case reflect.Map:
		return u.unmarshalMap(data, rv)
	case reflect.Array:
		return u.unmarshalArray(data, rv)
	case reflect.Slice:
		return u.unmarshalSlice(data, rv)
	case reflect.String:
		return u.unmarshalString(data, rv)
	case reflect.Interface:
		if rv.NumMethod() > 0 {
			return unsupported(rv.Kind())
		}
		return u.unmarshalAnything(data, rv)
	case reflect.Float32, reflect.Float64:
		return u.unmarshalFloat(data, rv)
	case reflect.Bool:
		return u.unmarshalBool(data, rv)
	}
	return unsupported(rv.Kind())
}

func (u unmarshaler) unmarshalStruct(data interface{}, rv reflect.Value) error {
	// fmt.Println("unmarshalStruct:", data, rv)
	mapping, ok := data.(map[string]interface{})
	if !ok {
		return mismatch(rv, "map", data)
	}

	fields := cachedTypeFields(rv.Type())
	for key, value := range mapping {
		// fmt.Println("unmarshalStruct.field:", key, value)
		f := fields[key]
		if f == nil {
			continue
		}
		sv := rv
		for j, i := range f.index {
			if j < len(f.index)-1 {
				sv = indirect(sv.Field(i))
			} else {
				sv = sv.Field(i)
			}
		}
		var stop bool
		var err error
		if value, stop, err = u.invokeHock(value, sv); err != nil {
			return err
		} else if stop {
			continue
		} else {
			sv = indirect(sv)
		}
		if unifiable(sv) {
			if err := u.unmarshal(value, sv); err != nil {
				return fmt.Errorf("unmarshal field %s fail: %s", key, err.Error())
			}
		} else {
			return fmt.Errorf("Field '%s.%s' is unexported, and therefore cannot be loaded with reflection.", rv.Type().String(), key)
		}
	}
	return nil
}

func (u unmarshaler) unmarshalSliceMap(data interface{}, rv reflect.Value) error {
	var items []struct {
		Key   interface{}
		Value interface{}
	}
	if err := u.unmarshal(data, rvalue(&items)); err != nil {
		return err
	}
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}
	for _, item := range items {
		rvKey := indirect(reflect.New(rv.Type().Key()))
		rvValue := indirect(reflect.New(rv.Type().Elem()))
		if err := u.unmarshal(item.Key, rvKey); err != nil {
			return fmt.Errorf("invalid key: %s", err.Error())
		}
		if err := u.unmarshal(item.Value, rvValue); err != nil {
			return fmt.Errorf("invalid value: %s", err.Error())
		}
		rv.SetMapIndex(rvKey, rvValue)
	}
	return nil
}

func (u unmarshaler) unmarshalMap(data interface{}, rv reflect.Value) error {
	mapping, ok := data.(map[string]interface{})
	if !ok {
		return u.unmarshalSliceMap(data, rv)
	}
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}
	for key, value := range mapping {
		rvval := reflect.Indirect(reflect.New(rv.Type().Elem()))
		if err := u.unmarshal(value, rvval); err != nil {
			return err
		}
		rv.SetMapIndex(reflect.ValueOf(key), rvval)
	}
	return nil
}

func (u unmarshaler) unmarshalArray(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	if lv.Kind() != reflect.Slice {
		return badtype("slice", data)
	}
	if lv.Len() != rv.Len() {
		return fmt.Errorf("Expected array length %d but got array of length %d",
			rv.Len(), lv.Len())
	}
	return u.unmarshalSliceArray(lv, rv)
}

func (u unmarshaler) unmarshalSlice(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	if lv.Kind() != reflect.Slice {
		return badtype("slice", data)
	}
	if rv.IsNil() {
		rv.Set(reflect.MakeSlice(rv.Type(), lv.Len(), lv.Len()))
	}
	return u.unmarshalSliceArray(lv, rv)
}

func (u unmarshaler) unmarshalSliceArray(lv, rv reflect.Value) error {
	for i, sliceLen := 0, lv.Len(); i < sliceLen; i++ {
		if value, stop, err := u.invokeHock(lv.Index(i).Interface(), rv.Index(i)); err != nil {
			return err
		} else if stop {
			continue
		} else if err := u.unmarshal(value, indirect(rv.Index(i))); err != nil {
			return err
		}
	}
	return nil
}

func (u unmarshaler) unmarshalString(data interface{}, rv reflect.Value) error {
	if s, ok := data.(string); ok {
		rv.SetString(s)
		return nil
	}
	return badtype("string", data)
}

func (u unmarshaler) unmarshalTime(data interface{}, rv reflect.Value) error {
	switch v := data.(type) {
	case time.Time:
		rv.Set(reflect.ValueOf(v))
	case string:
		var t time.Time
		if err := t.UnmarshalText([]byte(v)); err != nil {
			return fmt.Errorf("bad time format: %s", err.Error())
		}
		rv.Set(reflect.ValueOf(t))
	}
	return badtype("time.Time", data)
}

func (u unmarshaler) unmarshalDuration(data interface{}, rv reflect.Value) error {
	switch v := data.(type) {
	case time.Duration:
		rv.Set(reflect.ValueOf(v))
	case string:
		if d, err := time.ParseDuration(v); err != nil {
			return fmt.Errorf("bad duration format: %s", err.Error())
		} else {
			rv.Set(reflect.ValueOf(d))
		}
	default:
		return badtype("time.Duration", data)
	}
	return nil
}

func (u unmarshaler) unmarshalFloat(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	lk := lv.Kind()
	switch {
	case lk >= reflect.Int && lk <= reflect.Int64:
		rv.SetFloat(float64(lv.Int()))
	case lk >= reflect.Uint && lk <= reflect.Uint64:
		rv.SetFloat(float64(lv.Uint()))
	case lk == reflect.Float32 || lk == reflect.Float64:
		rv.SetFloat(lv.Float())
	default:
		return badtype("float", data)
	}
	return nil
}

func (u unmarshaler) unmarshalInt(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	lk := lv.Kind()
	rk := rv.Kind()
	if rk >= reflect.Int && rk <= reflect.Int64 {
		switch {
		case lk >= reflect.Int && lk <= reflect.Int64:
			rv.SetInt(lv.Int())
		case lk >= reflect.Uint && lk <= reflect.Uint64:
			rv.SetInt(int64(lv.Uint()))
		default:
			return badtype("integer", data)
		}
	} else if rk >= reflect.Uint && rk <= reflect.Uint64 {
		switch {
		case lk >= reflect.Int && lk <= reflect.Int64:
			rv.SetUint(uint64(lv.Int()))
		case lk >= reflect.Uint && lk <= reflect.Uint64:
			rv.SetUint(lv.Uint())
		default:
			return badtype("integer", data)
		}
	} else {
		return badtype("integer", data)
	}
	return nil
}

func (u unmarshaler) unmarshalBool(data interface{}, rv reflect.Value) error {
	if b, ok := data.(bool); ok {
		rv.SetBool(b)
		return nil
	}
	return badtype("boolean", data)
}

func (u unmarshaler) unmarshalAnything(data interface{}, rv reflect.Value) error {
	rv.Set(reflect.ValueOf(data))
	return nil
}

// copy from BurntSushi/toml
func (u unmarshaler) unmarshalText(data interface{}, v encoding.TextUnmarshaler) error {
	var s string
	switch sdata := data.(type) {
	case encoding.TextMarshaler:
		text, err := sdata.MarshalText()
		if err != nil {
			return err
		}
		s = string(text)
	case fmt.Stringer:
		s = sdata.String()
	case string:
		s = sdata
	case bool:
		s = fmt.Sprintf("%v", sdata)
	case int64:
		s = fmt.Sprintf("%d", sdata)
	case float64:
		s = fmt.Sprintf("%f", sdata)
	default:
		return badtype("primitive (string-like)", data)
	}
	if err := v.UnmarshalText([]byte(s)); err != nil {
		return err
	}
	return nil
}

// copy from BurntSushi/toml
func badtype(expected string, data interface{}) error {
	return fmt.Errorf("Expected %s but found '%T'.", expected, data)
}

// copy from BurntSushi/toml
func mismatch(user reflect.Value, expected string, data interface{}) error {
	return fmt.Errorf("Type mismatch for %s. Expected %s but found '%T'.",
		user.Type().String(), expected, data)
}

func unsupported(k reflect.Kind) error {
	return fmt.Errorf("Unsupported type \"%s\".", k)
}

func unifiable(rv reflect.Value) bool {
	if rv.CanSet() {
		return true
	}
	if _, ok := rv.Interface().(encoding.TextUnmarshaler); ok {
		return true
	}
	return false
}

// rvalue returns a reflect.Value of `v`. All pointers are resolved.
// copy from BurntSushi/toml
func rvalue(v interface{}) reflect.Value {
	return indirect(reflect.ValueOf(v))
}

// indirect returns the value pointed to by a pointer.
// Pointers are followed until the value is not a pointer.
// New values are allocated for each nil pointer.
//
// An exception to this rule is if the value satisfies an interface of
// interest to us (like encoding.TextUnmarshaler).
// copy from BurntSushi/toml
func indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		if v.CanAddr() {
			pv := v.Addr()
			if _, ok := pv.Interface().(encoding.TextUnmarshaler); ok {
				return pv
			}
		}
		return v
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return indirect(reflect.Indirect(v))
}

func (u unmarshaler) invokeHock(data interface{}, rv reflect.Value) (interface{}, bool, error) {
	if u.hock == nil {
		return data, false, nil
	} else {
		return u.hock(data, rv)
	}
}
