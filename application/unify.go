package application

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

func (app *Application) unify(data interface{}, rv reflect.Value) error {
	if v, ok := data.(string); ok && strings.HasPrefix(v, "!REF:") {
		if name := v[5:]; name == "" {
			return errors.New("Missing ref name.")
		} else {
			rv.Set(reflect.ValueOf(<-app.getModule(name)))
			return nil
		}
	}
	if v, ok := rv.Addr().Interface().(encoding.TextUnmarshaler); ok {
		return app.unifyText(data, v)
	}
	if rv.Type() == timeType {
		return app.unifyTime(data, rv)
	}
	if rv.Type() == durationType {
		return app.unifyDuration(data, rv)
	}

	k := rv.Kind()

	if k >= reflect.Int && k <= reflect.Uint64 {
		return app.unifyInt(data, rv)
	}
	switch k {
	case reflect.Ptr:
		elem := reflect.New(rv.Type().Elem())
		if err := app.unify(data, reflect.Indirect(elem)); err != nil {
			return err
		}
		rv.Set(elem)
		return nil
	case reflect.Struct:
		return app.unifyStruct(data, rv)
	case reflect.Map:
		return app.unifyMap(data, rv)
	case reflect.Array:
		return app.unifyArray(data, rv)
	case reflect.Slice:
		return app.unifySlice(data, rv)
	case reflect.String:
		return app.unifyString(data, rv)
	case reflect.Interface:
		if rv.NumMethod() > 0 {
			return unsupported(rv.Kind())
		}
		return app.unifyAnything(data, rv)
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return app.unifyFloat(data, rv)
	}
	return unsupported(rv.Kind())
}

func (app *Application) unifyStruct(data interface{}, rv reflect.Value) error {
	mapping, ok := data.(map[string]interface{})
	if !ok {
		return mismatch(rv, "map", mapping)
	}

	fields := cachedTypeFields(rv.Type())
	for key, value := range mapping {
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
		if s, ok := value.(string); ok && strings.HasPrefix(s, "!REF:") {
			if name := s[5:]; name == "" {
				return errors.New("Missing ref name.")
			} else {
				sv.Set(reflect.ValueOf(<-app.getModule(name)))
				continue
			}
		} else {
			sv = indirect(sv)
		}
		if unifiable(sv) {
			if err := app.unify(value, sv); err != nil {
				return fmt.Errorf("Type mismatch for \"%s.%s\": %s",
					rv.Type().String(), key, err.Error())
			}
		} else {
			return fmt.Errorf("Field '%s.%s' is unexported, and therefore cannot be loaded with reflection.", rv.Type().String(), key)
		}
	}
	return nil
}

func (app *Application) unifyMap(data interface{}, rv reflect.Value) error {
	mapping, ok := data.(map[string]interface{})
	if !ok {
		return badtype("map", mapping)
	}
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}
	for key, value := range mapping {
		rvval := reflect.Indirect(reflect.New(rv.Type().Elem()))
		if err := app.unify(value, rvval); err != nil {
			return err
		}
		rv.SetMapIndex(reflect.ValueOf(key), rvval)
	}
	return nil
}

func (app *Application) unifyArray(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	if lv.Kind() != reflect.Slice {
		return badtype("slice", data)
	}
	if lv.Len() != rv.Len() {
		return fmt.Errorf("Expected array length %d but got array of length %d",
			rv.Len(), lv.Len())
	}
	return app.unifySliceArray(lv, rv)
}

func (app *Application) unifySlice(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	if lv.Kind() != reflect.Slice {
		return badtype("slice", data)
	}
	if rv.IsNil() {
		rv.Set(reflect.MakeSlice(rv.Type(), lv.Len(), lv.Len()))
	}
	return app.unifySliceArray(lv, rv)
}

func (app *Application) unifySliceArray(lv, rv reflect.Value) error {
	for i, sliceLen := 0, lv.Len(); i < sliceLen; i++ {
		if s, ok := lv.Index(i).Interface().(string); ok {
			if strings.HasPrefix(s, "!REF:") {
				if name := s[5:]; name == "" {
					return errors.New("Missing ref name.")
				} else {
					rv.Index(i).Set(reflect.ValueOf(<-app.getModule(name)))
					continue
				}
			}
		}
		if err := app.unify(lv.Index(i).Interface(), indirect(rv.Index(i))); err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) unifyString(data interface{}, rv reflect.Value) error {
	if s, ok := data.(string); ok {
		rv.SetString(s)
		return nil
	}
	return badtype("string", data)
}

func (app *Application) unifyTime(data interface{}, rv reflect.Value) error {
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

func (app *Application) unifyDuration(data interface{}, rv reflect.Value) error {
	switch v := data.(type) {
	case time.Duration:
		rv.Set(reflect.ValueOf(v))
	case string:
		if d, err := time.ParseDuration(v); err != nil {
			return fmt.Errorf("bad duration format: %s", err.Error())
		} else {
			rv.Set(reflect.ValueOf(d))
		}
	}
	return badtype("time.Duration", data)
}

func (app *Application) unifyFloat(data interface{}, rv reflect.Value) error {
	lv := reflect.ValueOf(data)
	lk := lv.Kind()
	switch {
	case lk >= reflect.Int && lk <= reflect.Int64:
		rv.SetFloat(float64(lv.Int()))
	case lk >= reflect.Uint && lk <= reflect.Uint64:
		rv.SetFloat(float64(lv.Uint()))
	case lk == reflect.Float32 && lk == reflect.Float64:
		rv.SetFloat(lv.Float())
	default:
		return badtype("float", data)
	}
	return nil
}

func (app *Application) unifyInt(data interface{}, rv reflect.Value) error {
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

func (app *Application) unifyBool(data interface{}, rv reflect.Value) error {
	if b, ok := data.(bool); ok {
		rv.SetBool(b)
		return nil
	}
	return badtype("boolean", data)
}

func (app *Application) unifyAnything(data interface{}, rv reflect.Value) error {
	rv.Set(reflect.ValueOf(data))
	return nil
}

// copy from BurntSushi/toml
func (app *Application) unifyText(data interface{}, v encoding.TextUnmarshaler) error {
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
