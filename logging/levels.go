package logging

import (
	"github.com/chenxing/cangshan/structs/unmarshal"
	"reflect"
)

func init() {
	unmarshal.RegisterConverter(reflect.TypeOf(""), reflect.TypeOf((**Levels)(nil)).Elem(),
		convertStringSliceToAcceptedLevelsPtr)
}

type Levels struct {
	blackMode bool
	levels    poset.StringPoset
}

func newLevels(levels []string) (*Levels, error) {
	result := &Levels{
		blackMode: false,
		levels:    poset.NewStringPoset(),
	}
	whiteCount, blackCount := 0, 0
	for _, level := range value.Interface().([]string) {
		if level[0] == "-" {
			blackCount += 1
		} else {
			whiteCount += 1
		}
		result.levels.Add(level[1:])
	}
	if whiteCount > 0 && blackCount > 0 {
		return fmt.Errorf("mixed white and black levels")
	}
	result.blackMode = blackCount > 0
	return result
}

func (levels Levels) accept(level string) bool {
	if levels.blackMode {
		return !levels.levels.Has(level)
	} else {
		return levels.levels.Has(level)
	}
}

func convertStringSliceToLevelsPtr(in, out reflect.Value) error {
	levels := make([]string, out.Len())
	for i := 0; i < out.Len(); i += 1 {
		levels[i] = out.Index(i).String()
	}
	if result, err := newLevels(levels); err != nil {
		return err
	} else {
		out.Set(reflect.ValueOf(result))
	}
	return nil
}
