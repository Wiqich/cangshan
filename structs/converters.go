package unmarshaler

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

func init() {
	RegisterConverter(stringType, stringType, convertStringToString)
	RegisterConverter(stringType, reflect.TypeOf(int(0)), convertStringToInt)
	RegisterConverter(stringType, reflect.TypeOf(int8(0)), convertStringToInt)
	RegisterConverter(stringType, reflect.TypeOf(int16(0)), convertStringToInt)
	RegisterConverter(stringType, reflect.TypeOf(int32(0)), convertStringToInt)
	RegisterConverter(stringType, reflect.TypeOf(int64(0)), convertStringToInt)
	RegisterConverter(stringType, reflect.TypeOf(uint(0)), convertStringToUint)
	RegisterConverter(stringType, reflect.TypeOf(uint8(0)), convertStringToUint)
	RegisterConverter(stringType, reflect.TypeOf(uint16(0)), convertStringToUint)
	RegisterConverter(stringType, reflect.TypeOf(uint32(0)), convertStringToUint)
	RegisterConverter(stringType, reflect.TypeOf(uint64(0)), convertStringToUint)
	RegisterConverter(stringType, reflect.TypeOf(float32(0)), convertStringToFloat)
	RegisterConverter(stringType, reflect.TypeOf(float64(0)), convertStringToFloat)
	RegisterConverter(stringType, reflect.TypeOf(false), convertStringToBool)
}

type Converter func(reflect.Value, reflect.Value) error

func RegisterConverter(from reflect.Type, to reflect.Type, converter Converter) error {
	if converter == nil {
		return errors.New("converter cannot be nil")
	}
	converters[[2]reflect.Type{from, to}] = converter
	return nil
}

func convert(in, out reflect.Value) error {
	if converter := converters[[2]reflect.Type{in.Type(), out.Type()}]; converter == nil {
		return errNoConverter
	} else {
		return converter(in, out)
	}
}

var (
	converters     = make(map[[2]reflect.Type]Converter)
	errNoConverter = errors.New("no converter")
	hexPattern     = regexp.MustCompile("^0x[0-9a-fA-F]+$")
	octPattern     = regexp.MustCompile("^0[0-7]+$")
	dexPattern     = regexp.MustCompile("^-?[0-9]+$")
	stringType     = reflect.TypeOf("")
)

func convertStringToString(in, out reflect.Value) error {
	out.SetString(in.String())
	return nil
}

func convertStringToInt(in, out reflect.Value) error {
	if v, err := parseInt64(in.String()); err != nil {
		return err
	} else {
		in.SetInt(v)
	}
	return nil
}

func convertStringToUint(in, out reflect.Value) error {
	if v, err := parseUint64(in.String()); err != nil {
		return err
	} else {
		in.SetUint(v)
	}
	return nil
}

func convertStringToFloat(in, out reflect.Value) error {
	if v, err := strconv.ParseFloat(in.String(), 64); err != nil {
		return err
	} else {
		in.SetFloat(v)
	}
	return nil
}

func convertStringToBool(in, out reflect.Value) error {
	if v, err := strconv.ParseBool(in.String()); err != nil {
		return err
	} else {
		in.SetBool(v)
	}
	return nil
}

func parseInt64(text string) (int64, error) {
	if hexPattern.MatchString(text) {
		return strconv.ParseInt(text[2:], 16, 64)
	} else if octPattern.MatchString(text) {
		return strconv.ParseInt(text[1:], 8, 64)
	} else if dexPattern.MatchString(text) {
		return strconv.ParseInt(text, 10, 64)
	} else {
		return 0, fmt.Errorf("bad int format: %s", text)
	}
}

func parseUint64(text string) (uint64, error) {
	if hexPattern.MatchString(text) {
		return strconv.ParseUint(text[2:], 16, 64)
	} else if octPattern.MatchString(text) {
		return strconv.ParseUint(text[1:], 8, 64)
	} else if dexPattern.MatchString(text) && text[0] != '-' {
		return strconv.ParseUint(text, 10, 64)
	} else {
		return 0, fmt.Errorf("bad uint format: %s", text)
	}
}
