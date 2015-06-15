package unmarshaler

import (
	"fmt"
	"reflect"
	"strconv"
)

type unmarshalError struct {
	err  error
	path string
}

func newUnmarshalError(format string, params ...interface{}) *unmarshalError {
	return &unmarshalError{
		err:  fmt.Errorf(format, params...),
		path: "",
	}
}

func (err *unmarshalError) wrap(path string) *unmarshalError {
	if err.path != "" {
		if err.path[0] == '[' {
			path += err.path
		} else {
			path += "." + err.path
		}
	}
	return &unmarshalError{
		err:  err.err,
		path: path,
	}
}

func Unmarshal(in map[string]interface{}, out interface{}) error {
	err := unmarshal(reflect.ValueOf(in), reflect.ValueOf(out))
	if err.path == "" {
		return fmt.Errorf("unmarshal struct fail: %s", err.err.Error())
	} else {
		return fmt.Errorf("unmarshal struct field %s fail: %s", err.path, err.err.Error())
	}
}

func unmarshal(in, out reflect.Value) *unmarshalError {
	for out.Kind() == reflect.Ptr {
		out.Set(reflect.New(out.Elem().Type()))
		out = out.Elem()
	}
	inType := in.Type()
	inKind := inType.Kind()
	switch {
	case inKind == reflect.Bool:
		return unmarshalBool(in, out)
	case inKind >= reflect.Int && inKind <= reflect.Float64 && inKind != reflect.Uintptr:
		return unmarshalNum(in, out)
	case inKind == reflect.Map && inType.Key().Kind() == reflect.String:
		return unmarshalMap(in, out)
	case inKind == reflect.Slice:
		return unmarshalSlice(in, out)
	case inKind == reflect.String:
		return unmarshalString(in, out)
	default:
		return newUnmarshalError("unsupported in type: %s", inType)
	}
}

func unmarshalMap(in, out reflect.Value) *unmarshalError {
	out.Set(reflect.MakeMap(out.Type()))
	outType := out.Type()
	switch outType.Kind() {
	case reflect.Struct:
		return unmarshalStruct(in, out)
	case reflect.Map:
		if outType.Key() != nil {
			return newUnmarshalError("expect type map[string]interface{}, not %s", outType)
		}
		for _, key := range in.MapKeys() {
			value := in.MapIndex(key)
			outValue := reflect.New(outType.Elem())
			if err := unmarshal(value, outValue); err != nil {
				return err.wrap(fmt.Sprintf("[%s]", key.String()))
			}
			out.SetMapIndex(key, outValue.Elem())
		}
	}
	return newUnmarshalError("expect type struct/map[string]interface{}, not %s", outType)
}

func unmarshalStruct(in, out reflect.Value) *unmarshalError {
	outType := out.Type()
	inMap := make(map[string]reflect.Value)
	for _, key := range in.MapKeys() {
		inMap[key.String()] = in.MapIndex(key)
	}
	for i := 0; i < outType.NumField(); i += 1 {
		field := outType.Field(i)
		if inValue, found := inMap[field.Name]; !found {
			if defaultValue := field.Tag.Get("defaultValue"); defaultValue != "" {
				if err := setDefaultValue(out.Field(i), defaultValue); err != nil {
					return err.wrap(field.Name)
				}
			}
		}
		pluginDone := false
		for name, plugin := range unmarshalPlugins {
			if tag := field.Tag.Get(name); tag != "" {
				if err := plugin.Unmarshal(inValue, out.Field(i), tag); err == ErrIgnorePlugin {
					continue
				} else if err != nil {
					return newUnmarshalError(err.Error()).wrap(name)
				} else {
					pluginDone = true
					break
				}
			}
		}
		if !pluginDone {
			if err := unmarshal(inValue, out.Field(i)); err != nil {
				return err.wrap(name)
			}
		}
	}
	return nil
}

func unmarshalSlice(in, out reflect.Value) *unmarshalError {
	if out.Kind() != reflect.Slice {
		return newUnmarshalError("expect type []interface{}, not %s", out.Type())
	}
	out.Set(reflect.MakeSlice(out.Type(), in.Len(), in.Cap()))
	for i := 0; i < in.Len(); i += 1 {
		if err := unmarshal(in.Index(i), out.Index(i)); err != nil {
			return err.wrap(fmt.Sprintf("[%d]", i))
		}
	}
	return nil
}

func unmarshalString(in, out reflect.Value) *unmarshalError {
	if out.Kind() != reflect.String {
		return newUnmarshalError("expect type string, not %s", out.Type())
	}
	out.SetString(in.String())
	return nil
}

func unmarshalBool(in, out reflect.Value) *unmarshalError {
	if out.Kind() != reflect.String {
		return newUnmarshalError("expect type bool, not %s", out.Type())
	}
	out.SetBool(in.Bool())
	return nil
}

func unmarshalNum(in, out reflect.Value) *unmarshalError {
	if out.Kind() < reflect.Int || out.Kind() > reflect.Float64 {
		return newUnmarshalError("expect type int ~ float64, not %s", out.Type())
	}
	out.Set(in.Convert(out.Type()))
	return nil
}

func setDefaultValue(out reflect.Value, defaultValue string) *unmarshalError {
	outKind := out.Type().Kind()
	switch {
	case outKind == reflect.Bool:
		if value, err := strconv.ParseBool(defaultValue); err != nil {
			return newUnmarshalError("invalid bool default value: %s", err.Error())
		} else {
			out.SetBool(value)
		}
	case outKind >= reflect.Int && outKind <= reflect.Int64:
		var value int64
		var err error
		if defaultValue[0] == "0" {
			if len(defaultValue) > 2 && defaultValue[1] == "x" {
				value, err = strconv.ParseInt(defaultValue[2:], 16, 64)
			} else if len(defaultValue) > 1 {
				value, err = strconv.ParseInt(defaultValue[1:], 8, 64)
			} else {
				value = 0
			}
		} else {
			value, err = strconv.ParseInt(defaultValue, 10, 64)
		}
		if err != nil {
			return newUnmarshalError("invalid int default value: %s", defaultValue)
		}
		out.SetInt(value)
	case outKind >= reflect.Uint && outKind <= reflect.Uint64:
		var value uint64
		var err error
		if defaultValue[0] == "0" {
			if len(defaultValue) > 2 && defaultValue[1] == "x" {
				value, err = strconv.ParseUint(defaultValue[2:], 16, 64)
			} else if len(defaultValue) > 1 {
				value, err = strconv.ParseUint(defaultValue[1:], 8, 64)
			} else {
				value = 0
			}
		} else {
			value, err = strconv.ParseUint(defaultValue, 10, 64)
		}
		if err != nil {
			return newUnmarshalError("invalid uint default value: %s", defaultValue)
		}
		out.SetUint(value)
	case outKind == reflect.Float32 || outKind == reflect.Float64:
		if value, err := strconv.ParseFloat(defaultValue, 64); err != nil {
			return newUnmarshalError("invalid float default value: %s", defaultValue)
		} else {
			out.SetFloat(value)
		}
	// TODO: support simple slice
	default:
		return newUnmarshalError("unsupport default value type: %s", out.Type())
	}
}
