package jaml

import (
	"fmt"
	"reflect"
	"time"
)

var (
	flags = make(map[string]interface{})
)

func BoolFlag(name string, defaultValue bool) *bool {
	flag := new(bool)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func IntFlag(name string, defaultValue int) *int {
	flag := new(int)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func Int64Flag(name string, defaultValue int64) *int64 {
	flag := new(int64)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func Float64Flag(name string, defaultValue float64) *float64 {
	flag := new(float64)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func StringFlag(name string, defaultValue string) *string {
	flag := new(string)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func TimeFlag(name string, defaultValue time.Time) *time.Time {
	flag := new(time.Time)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func DurationFlag(name string, defaultValue time.Duration) *time.Duration {
	flag := new(time.Duration)
	*flag = defaultValue
	flags[name] = flag
	return flag
}

func setFlag(flag interface{}, value *Value) error {
	switch value.Kind {
	case reflect.Int64:
		switch v := flag.(type) {
		case *time.Duration:
			var err error
			if *v, err = time.ParseDuration(value.Primary.(string)); err != nil {
				return err
			}
			return nil
		case *int:
			*v = int(value.Primary.(int64))
			return nil
		case *int64:
			*v = value.Primary.(int64)
			return nil
		}
	case reflect.Float64:
		if v, ok := flag.(*float64); ok {
			*v = value.Primary.(float64)
			return nil
		}
	case reflect.Bool:
		if v, ok := flag.(*bool); ok {
			*v = value.Primary.(bool)
			return nil
		}
	case reflect.String:
		if v, ok := flag.(*string); ok {
			*v = value.Primary.(string)
			return nil
		}
	case reflect.Struct:
		switch value.Type {
		case "time.Time":
			if v, ok := flag.(*time.Time); ok {
				var err error
				if *v, err = parseTime(value.Primary.(string)); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return fmt.Errorf("flag type mismatch: %s vs. %t", value.Kind, flag)
}
