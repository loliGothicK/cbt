package cutil

import (
	"os"
	"reflect"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
func Apply(f interface{}, args ...interface{}) []interface{} {
	var rvs []reflect.Value
	for _, arg := range args {
		rvs = append(rvs, reflect.ValueOf(arg))
	}
	rvs = reflect.ValueOf(f).Call(rvs)
	var ret []interface{}
	for _, result := range rvs {
		ret = append(ret, result.Interface())
	}
	return ret
}
func OrElse(b bool, t, f interface{}) interface{} {
	switch {
	case b:
		return t
	default:
		return f
	}
}
