package unmarshaler

import (
	"fmt"
	"reflect"
	"strconv"
)

var (
	unmarshalHandlers map[reflect.Kind]func(reflect.Value, reflect.Value) *unmarshalError
)

func init() {
	unmarshalHandlers = map[reflect.Kind]func(reflect.Value, reflect.Value) *unmarshalError{
		reflect.Bool:    unmarshalBool,
		reflect.Int:     unmarshalNum,
		reflect.Int8:    unmarshalNum,
		reflect.Int16:   unmarshalNum,
		reflect.Int32:   unmarshalNum,
		reflect.Int64:   unmarshalNum,
		reflect.Uint:    unmarshalNum,
		reflect.Uint8:   unmarshalNum,
		reflect.Uint16:  unmarshalNum,
		reflect.Uint32:  unmarshalNum,
		reflect.Uint64:  unmarshalNum,
		reflect.Float32: unmarshalNum,
		reflect.Float64: unmarshalNum,
		reflect.Map:     unmarshalMap,
		reflect.Slice:   unmarshalSlice,
		reflect.String:  unmarshalString,
		reflect.Struct:  unmarshalStruct,
	}
}

func Unmarshal(in map[string]interface{}, out interface{}) error {
	return UnmarshalValue(reflect.ValueOf(in), reflect.ValueOf(out))
}

func UnmarshalValue(in, out reflect.Value) error {
	err := unmarshal(in, out)
	if err.path == "" {
		return fmt.Errorf("unmarshal fail: %s", err.err.Error())
	} else {
		return fmt.Errorf("unmarshal field %s fail: %s", err.path, err.err.Error())
	}
}

func unmarshal(in, out reflect.Value) *unmarshalError {
	if err := convert(in, out); err != errNoConverter {
		// first: Converter
		if err != nil {
			return newUnmarshalError("convert fail: %s", err.Error())
		}
	} else {
		// second: do unmarhsaling
		if out.Kind() == reflect.Ptr {
			out.Set(reflect.New(out.Elem().Type()))
			out = out.Elem()
		}
		if handler := unmarshalHandlers[in.Kind()]; handler == nil {
			return newUnmarshalError("unsupported in type: %s", in.Type())
		} else if err := handler(in, out); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalMap(in, out reflect.Value) *unmarshalError {
	out.Set(reflect.MakeMap(out.Type()))
	outType := out.Type()
	if outType.Key().Kind() != reflect.Struct {
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
	return nil
}

func unmarshalStruct(in, out reflect.Value) *unmarshalError {
	outType := out.Type()
	inMap := make(map[string]reflect.Value)
	for _, key := range in.MapKeys() {
		inMap[key.String()] = in.MapIndex(key)
	}
	for i := 0; i < outType.NumField(); i += 1 {
		field := outType.Field(i)
		inValue, found := inMap[field.Name]
		if !found {
			if defaultValue := field.Tag.Get("default"); defaultValue != "" {
				inValue = reflect.ValueOf(defaultValue)
			} else {
				continue
			}
		}
		if err := unmarshal(inValue, out.Field(i)); err != nil {
			return err.wrap(field.Name)
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
		if defaultValue[0] == '0' {
			if len(defaultValue) > 2 && defaultValue[1] == 'x' {
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
		if defaultValue[0] == '0' {
			if len(defaultValue) > 2 && defaultValue[1] == 'x' {
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
	return nil
}
