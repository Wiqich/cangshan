package sqlresource

import (
	"database/sql"
	"fmt"
	"github.com/yangchenxing/cangshan/application"
	cssql "github.com/yangchenxing/cangshan/client/sql"
	"reflect"
	"strconv"
	"time"
)

func init() {
	application.RegisterBuiltinModule("SimpleREST.SQLResource.Int64Type", new(Int64Type))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.Float64Type", new(Float64Type))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.BoolType", new(BoolType))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.StringType", new(StringType))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.NullInt64Type", new(NullInt64Type))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.NullFloat64Type", new(NullFloat64Type))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.NullBoolType", new(NullBoolType))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.NullStringType", new(NullStringType))
	application.RegisterBuiltinModule("SimpleREST.SQLResource.DefaultNullTimeType", new(NullTime))
	application.RegisterModulePrototype("SimpleRESTSQLResourceNullTimeType", new(NullTime))
}

type Int64Type struct{}

func (t Int64Type) ValueHolder() interface{} {
	return new(int64)
}

func (t Int64Type) Decode(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case int, int8, int16, int32:
		return reflect.ValueOf(i).Int(), nil
	case int64:
		return v, nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(i).Uint()), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t Int64Type) Encode(i interface{}) interface{} {
	return *(i.(*int64))
}

type Float64Type struct{}

func (t Float64Type) ValueHolder() interface{} {
	return new(float64)
}

func (t Float64Type) Decode(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(i).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(i).Uint()), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t Float64Type) Encode(i interface{}) interface{} {
	return *(i.(*float64))
}

type BoolType struct{}

func (t BoolType) ValueHolder() interface{} {
	return new(bool)
}

func (t BoolType) Decode(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t BoolType) Encode(i interface{}) interface{} {
	return *(i.(*bool))
}

type StringType struct{}

func (t StringType) ValueHolder() interface{} {
	return new(string)
}

func (t StringType) Decode(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case string:
		return v, nil
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t StringType) Encode(i interface{}) interface{} {
	return *(i.(*string))
}

// Nullable types

type NullInt64Type struct{}

func (t NullInt64Type) ValueHolder() interface{} {
	return new(sql.NullInt64)
}

func (t NullInt64Type) Decode(i interface{}) (interface{}, error) {
	if i == nil {
		return nil, nil
	}
	switch v := i.(type) {
	case int, int8, int16, int32:
		return reflect.ValueOf(i).Int(), nil
	case int64:
		return v, nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(i).Uint()), nil
	case string:
		if v == "null" || v == "nil" {
			return nil, nil
		}
		return strconv.ParseInt(v, 10, 64)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t NullInt64Type) Encode(i interface{}) interface{} {
	v := *(i.(*sql.NullInt64))
	if !v.Valid {
		return nil
	}
	return v.Int64
}

type NullFloat64Type struct{}

func (t NullFloat64Type) ValueHolder() interface{} {
	return new(sql.NullFloat64)
}

func (t NullFloat64Type) Decode(i interface{}) (interface{}, error) {
	if i == nil {
		return nil, nil
	}
	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(i).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(i).Uint()), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		if v == "null" || v == "nil" {
			return nil, nil
		}
		return strconv.ParseFloat(v, 64)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t NullFloat64Type) Encode(i interface{}) interface{} {
	v := *(i.(*sql.NullFloat64))
	if !v.Valid {
		return nil
	}
	return v.Float64
}

type NullBoolType struct{}

func (t NullBoolType) ValueHolder() interface{} {
	return new(sql.NullBool)
}

func (t NullBoolType) Decode(i interface{}) (interface{}, error) {
	if i == nil {
		return nil, nil
	}
	switch v := i.(type) {
	case bool:
		return v, nil
	case string:
		if v == "null" || v == "nil" {
			return nil, nil
		}
		return strconv.ParseBool(v)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t NullBoolType) Encode(i interface{}) interface{} {
	v := *(i.(*sql.NullBool))
	if !v.Valid {
		return nil
	}
	return v.Bool
}

type NullStringType struct{}

func (t NullStringType) ValueHolder() interface{} {
	return new(sql.NullString)
}

func (t NullStringType) Decode(i interface{}) (interface{}, error) {
	if i == nil {
		return nil, nil
	}
	switch v := i.(type) {
	case string:
		if v == "null" || v == "nil" {
			return nil, nil
		}
		return v, nil
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t NullStringType) Encode(i interface{}) interface{} {
	v := *(i.(*sql.NullString))
	if !v.Valid {
		return nil
	}
	return v.String
}

// Date and time types

type NullTime struct {
	Format string
}

func (t NullTime) ValueHolder() interface{} {
	return new(cssql.NullTime)
}

func (t NullTime) Decode(i interface{}) (interface{}, error) {
	if i == nil {
		return nil, nil
	}
	switch v := i.(type) {
	case time.Time:
		return v, nil
	case string:
		if v == "null" || v == "nil" {
			return nil, nil
		}
		if t.Format != "" {
			return time.Parse(t.Format, v)
		}
		switch len(v) {
		case 8:
			return time.Parse("20060102", v)
		case 10:
			return time.Parse("2006-01-02", v)
		case 14:
			return time.Parse("20060102150405", v)
		case 19:
			return time.Parse("2006-01-02:15:04:05", v)
		}
		return nil, fmt.Errorf("Invalid date value: %s", v)
	}
	return nil, fmt.Errorf("Unsupported type: %t", i)
}

func (t NullTime) Encode(i interface{}) interface{} {
	v := *(i.(*cssql.NullTime))
	if !v.Valid {
		return nil
	}
	if t.Format != "" {
		return v.Time.Format(t.Format)
	}
	return v.Time
}
