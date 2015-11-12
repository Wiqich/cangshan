package jaml

import (
	"bytes"
	"encoding"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	timeLayouts = []string{
		"20060102150405",
		"2006-01-02:15:04:05",
		"20060102150405-0700",
		"2006-01-02:15:04:05-0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	}
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Second)
)

func parseTime(s string) (t time.Time, err error) {
	for _, layout := range timeLayouts {
		if len(layout) != len(s) {
			continue
		}
		if t, err = time.Parse(layout, s); err == nil {
			return
		}
	}
	err = fmt.Errorf("unknown time layout: %s", s)
	return
}

type MapItem struct {
	Key   *Value
	Value *Value
}

type Value struct {
	Kind     reflect.Kind      `json:"kind"`
	Type     string            `json:"type,omitempty"`
	Primary  interface{}       `json:"primary,omitempty"`
	Items    []*Value          `json:"items,omitempty"`
	MapItems []*MapItem        `json:"mapitems,omitempty"`
	Fields   map[string]*Value `json:"fields,omitempty"`
	Imports  []string          `json:"imports,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	lexer    *lexer
}

func (val Value) String() string {
	switch val.Kind {
	case reflect.Int64:
		return strconv.FormatInt(val.Primary.(int64), 64)
	case reflect.Float64:
		return strconv.FormatFloat(val.Primary.(float64), 'f', -1, 64)
	case reflect.String:
		return strconv.Quote(val.Primary.(string))
	case reflect.Bool:
		return strconv.FormatBool(val.Primary.(bool))
	case reflect.Slice:
		var buf bytes.Buffer
		buf.WriteRune('[')
		for i, item := range val.Items {
			if i != 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(item.String())
		}
		buf.WriteRune(']')
		return buf.String()
	case reflect.Map:
		var buf bytes.Buffer
		buf.WriteRune('{')
		for i, item := range val.MapItems {
			if i != 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(item.Key.String())
			buf.WriteRune(':')
			buf.WriteString(item.Value.String())
		}
		buf.WriteRune('}')
		return buf.String()
	case reflect.Struct:
		var buf bytes.Buffer
		buf.WriteRune('{')
		first := true
		for k, v := range val.Fields {
			if first {
				first = false
			} else {
				buf.WriteRune(',')
			}
			buf.WriteString(k)
			buf.WriteRune(':')
			buf.WriteString(v.String())
		}
		buf.WriteRune('}')
		return buf.String()
	case reflect.Ptr:
		return fmt.Sprintf("@%s", val.Primary.(string))
	}
	return fmt.Sprintf("Kind:%s", val.Kind.String())
}

func (val Value) Equal(v *Value) bool {
	if val.Kind != v.Kind {
		return false
	}
	switch val.Kind {
	case reflect.Int64:
		return val.Primary.(int64) == v.Primary.(int64)
	case reflect.Float64:
		return val.Primary.(float64) == v.Primary.(float64)
	case reflect.String:
		return val.Primary.(string) == v.Primary.(string)
	case reflect.Bool:
		return val.Primary.(bool) == v.Primary.(bool)
	}
	// 暂不支持其他类型
	return false
}

func (val *Value) Append(p string, v *Value) error {
	field, err := val.GetField(p)
	if err != nil {
		return err
	}
	switch field.Kind {
	case reflect.Invalid:
		field.Kind = reflect.Slice
		field.Items = make([]*Value, 0, 4)
	case reflect.Slice:
		break
	default:
		return fmt.Errorf("invalid kind of %s: %s", p, field.Kind)
	}
	field.Items = append(field.Items, v)
	return nil
}

func (val *Value) SetIndex(p string, k *Value, v *Value) error {
	field, err := val.GetField(p)
	if err != nil {
		return err
	}
	switch field.Kind {
	case reflect.Invalid:
		field.Kind = reflect.Map
		field.MapItems = make([]*MapItem, 0, 4)
	case reflect.Map:
		break
	default:
		return fmt.Errorf("invalid kind of %s: %s", p, field.Kind)
	}
	for _, item := range field.MapItems {
		if item.Key.Equal(k) {
			item.Value = v
			return nil
		}
	}
	field.MapItems = append(field.MapItems, &MapItem{k, v})
	return nil
}

func (val *Value) GetField(p string) (*Value, error) {
	if p == "" {
		return val, nil
	}
	f := val
	fp := strings.Split(p, ".")
	for i, p := range fp {
		switch f.Kind {
		case reflect.Invalid:
			f.Kind = reflect.Struct
			f.Fields = make(map[string]*Value)
		case reflect.Struct:
			break
		default:
			return nil, fmt.Errorf("invalid kind of %s: %s",
				strings.Join(fp[:i+1], "."), f.Kind)
		}
		ff, found := f.Fields[p]
		if !found {
			v := &Value{}
			f.Fields[p] = v
			ff = v
		}
		f = ff
	}
	return f, nil
}

func (val *Value) SetField(p string, v *Value) error {
	var err error
	if p == "" {
		return fmt.Errorf("invalid path: %q", p)
	}
	obj := val
	name := p
	pos := strings.LastIndex(p, ".")
	if pos != -1 {
		if obj, err = val.GetField(p[:pos]); err != nil {
			return err
		}
		name = p[pos+1:]
	}
	switch obj.Kind {
	case reflect.Invalid:
		obj.Kind = reflect.Struct
		obj.Fields = make(map[string]*Value)
	case reflect.Struct:
		break
	default:
		return fmt.Errorf("invalid kind of %s: %s", p, obj.Kind)
	}
	obj.Fields[name] = v
	return nil
}

func (val *Value) Import(p string) error {
	switch val.Kind {
	case reflect.Struct:
		if val.Imports == nil {
			val.Imports = make([]string, 0, 1)
		}
	case reflect.Invalid:
		val.Kind = reflect.Struct
		val.Imports = make([]string, 0, 1)
	default:
		return fmt.Errorf("invalid kind for import: %s", val.Kind)
	}
	val.Imports = append(val.Imports, p)
	return nil
}

func (val *Value) DoImport(dir string, debug func(string)) error {
	if val.Kind == reflect.Struct {
		for _, pth := range val.Imports {
			v, err := ParseFile(filepath.Join(dir, pth), debug)
			if err != nil {
				return err
			}
			for name, value := range v.Fields {
				val.SetField(name, value)
			}
		}
		for _, value := range val.Fields {
			value.DoImport(dir, debug)
		}
	}
	return nil
}

func (val *Value) Unmarshal(v reflect.Value, name string, deps chan<- moduleDepInfo) error {
	if val.Kind == reflect.String {
		if v.CanAddr() {
			if tu, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
				tu.UnmarshalText([]byte(val.Primary.(string)))
			}
		}
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Kind == reflect.Int32 && v.Kind() == reflect.Int32 {
			v.SetInt(int64(val.Primary.(int32)))
			break
		}
		if val.Kind != reflect.Int64 {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		if v.Type() == durationType {
			if d, err := time.ParseDuration(val.Primary.(string)); err != nil {
				return err
			} else {
				v.SetInt(int64(d))
			}
		} else {
			v.SetInt(val.Primary.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Kind != reflect.Int64 {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		v.SetUint(uint64(val.Primary.(int64)))
	case reflect.Float32, reflect.Float64:
		if val.Kind != reflect.Float64 {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		v.SetFloat(val.Primary.(float64))
	case reflect.Bool:
		if val.Kind != reflect.Bool {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		v.SetBool(val.Primary.(bool))
	case reflect.String:
		if val.Kind != reflect.String {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		v.SetString(val.Primary.(string))
	case reflect.Slice:
		if val.Kind == reflect.String && v.Type().Elem().Kind() == reflect.Uint8 {
			v.Set(reflect.ValueOf([]byte(val.Primary.(string))))
			break
		}
		if val.Kind != reflect.Slice {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type().Elem(), len(val.Items), len(val.Items)))
		}
		for i, item := range val.Items {
			if err := item.Unmarshal(v.Index(i), name, deps); err != nil {
				return fmt.Errorf("slice item %d error: %s", i, err.Error())
			}
		}
	case reflect.Array:
		if val.Kind != reflect.Slice {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		for i, item := range val.Items {
			if i >= v.Len() {
				break
			}
			if err := item.Unmarshal(v.Index(i), name, deps); err != nil {
				return fmt.Errorf("slice item %d error: %s", i, err.Error())
			}
		}
	case reflect.Map:
		if val.Kind != reflect.Map {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		key := reflect.New(v.Type().Key()).Elem()
		elem := reflect.New(v.Type().Elem()).Elem()
		for i, item := range val.MapItems {
			if err := item.Key.Unmarshal(key, name, deps); err != nil {
				return fmt.Errorf("map item %d key error: %s", i, err.Error())
			}
			if err := item.Value.Unmarshal(elem, name, deps); err != nil {
				return fmt.Errorf("map item %d value error: %s", i, err.Error())
			}
			v.SetMapIndex(key, elem)
		}
	case reflect.Ptr:
		if val.Kind == reflect.Ptr {
			ch := make(chan interface{})
			deps <- moduleDepInfo{name, val.Primary.(string), ch}
			mod := <-ch
			v.Set(reflect.ValueOf(mod))
		} else {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			if err := val.Unmarshal(reflect.Indirect(v), name, deps); err != nil {
				return err
			}
		}
	case reflect.Struct:
		if val.Kind != reflect.Struct {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		typ := v.Type()
		if typ == timeType {
			if val.Type != "time.Time" {
				return fmt.Errorf("type mismatch: \"time.Time\" vs. %q", val.Type)
			}
			if t, err := parseTime(val.Primary.(string)); err != nil {
				return err
			} else {
				v.Set(reflect.ValueOf(t))
			}
		} else {
			for fieldName, fieldValue := range val.Fields {
				if _, found := typ.FieldByName(fieldName); found {
					if err := fieldValue.Unmarshal(v.FieldByName(fieldName), name, deps); err != nil {
						return fmt.Errorf("field %s erro: %s", fieldName, err)
					}
				}
			}
		}
	case reflect.Interface:
		if val.Kind != reflect.Struct {
			return fmt.Errorf("kind mismatch: %s vs. %s", v.Kind(), val.Kind)
		}
		obj := create(val.Type)
		if obj == nil {
			return fmt.Errorf("unknown type: %q", val.Type)
		}
		if !reflect.TypeOf(obj).Implements(v.Type()) {
			return fmt.Errorf("type %q don't implement %q", val.Type, v.Type().Name())
		}
		v.Set(reflect.ValueOf(obj))
		if err := val.Unmarshal(v.Elem(), name, deps); err != nil {
			return err
		}
	}
	return nil
}
