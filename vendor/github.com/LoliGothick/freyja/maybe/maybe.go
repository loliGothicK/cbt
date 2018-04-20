package maybe

import (
	"fmt"
	"reflect"
)

type Maybe struct {
	value interface{}
	err   error
}

func Some(x interface{}) Maybe {
	return Maybe{x, nil}
}

func None() Maybe {
	return Maybe{nil, fmt.Errorf(`java.lang.NullPointerException`)}
}

func Expected(ret ...interface{}) Maybe {
	if len(ret) != 2 {
		panic(`Error: Invalid arguments.`)
	}
	if ret[1] == nil {
		return Maybe{ret[0], nil}
	} else {
		return Maybe{nil, ret[1].(error)}
	}
}

func (maybe Maybe) Map(f interface{}) Maybe {
	if fv := reflect.ValueOf(f); fv.Kind() != reflect.Func {
		panic("argument is not func.")
	} else if maybe.err != nil {
		return Maybe{nil, maybe.err}
	} else {
		return Maybe{fv.Call([]reflect.Value{reflect.ValueOf(maybe.value)})[0].Interface(), nil}
	}
}

func (maybe Maybe) MapOr(dft interface{}, f interface{}) interface{} {
	if fv := reflect.ValueOf(f); fv.Kind() != reflect.Func {
		panic("2nd argument is not func.")
	} else if maybe.err != nil {
		return dft
	} else {
		return fv.Call([]reflect.Value{reflect.ValueOf(maybe.value)})[0].Interface()
	}
}

func (maybe Maybe) MapOrElse(d interface{}, f interface{}) interface{} {
	if fv := reflect.ValueOf(f); fv.Kind() != reflect.Func {
		panic("2nd argument is not func.")
	} else if maybe.err != nil {
		if dv := reflect.ValueOf(d); fv.Kind() != reflect.Func {
			panic("1st argument is not func.")
		} else {
			return dv.Call([]reflect.Value{})[0].Interface()
		}
	} else {
		return fv.Call([]reflect.Value{reflect.ValueOf(maybe.value)})[0].Interface()
	}
}

func (maybe Maybe) IsSome() bool {
	return maybe.value != nil
}
func (maybe Maybe) IsNone() bool {
	return maybe.value == nil
}
func (maybe Maybe) Expect(msg string) interface{} {
	if maybe.value != nil {
		return maybe.value
	} else {
		panic(msg)
	}
}
func (maybe Maybe) Unwrap() interface{} {
	if maybe.value != nil {
		return maybe.value
	} else {
		panic(maybe.err)
	}
}
func (maybe Maybe) UnwrapOr(or interface{}) interface{} {
	if maybe.value != nil {
		return maybe.value
	} else {
		return or
	}
}
func (maybe Maybe) UnwrapOrElse(f interface{}) interface{} {
	if maybe.value != nil {
		return maybe.value
	} else {
		fv := reflect.ValueOf(f)
		if fv.Kind() != reflect.Func {
			panic("1st argument is not func.")
		}
		return fv.Call([]reflect.Value{})[0].Interface()
	}
}
func (x Maybe) And(y Maybe) Maybe {
	switch {
	case x.value != nil:
		return y
	default:
		return x
	}
}
func (maybe Maybe) AndThen(f interface{}) Maybe {
	if fv := reflect.ValueOf(f); fv.Kind() != reflect.Func {
		panic("argument is not func.")
	} else if maybe.err != nil {
		return Maybe{nil, maybe.err}
	} else if reflect.TypeOf(f).In(0) == reflect.TypeOf(f).Out(0) {
		if result := fv.Call([]reflect.Value{reflect.ValueOf(maybe.value)}); result[1].Interface() == nil {
			return Maybe{result[0].Interface(), nil}
		} else {
			return Maybe{nil, result[1].Interface().(error)}
		}
	} else {
		panic(`Invalid func.`)
	}
}

func (maybe Maybe) OkOr(e error) interface{} {
	switch {
	case maybe.value != nil:
		return maybe.value
	default:
		return e
	}
}

func (maybe Maybe) OkOrElse(f interface{}) interface{} {
	switch {
	case maybe.value != nil:
		return maybe.value
	default:
		return reflect.ValueOf(f).Call([]reflect.Value{})[0].Interface()
	}
}

func (maybe Maybe) Interface() interface{} {
	switch {
	case maybe.value != nil:
		return maybe.value
	default:
		return maybe.err
	}
}
