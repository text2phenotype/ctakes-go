package maps

import (
	"fmt"
	"reflect"
)

type structValue struct {
	reflect.Value
}
type pointerValue struct {
	reflect.Value
}

func asPointer(v reflect.Value) pointerValue {
	return pointerValue{v}
}

func asStruct(v reflect.Value) structValue {
	return structValue{v}
}

func (v structValue) pointer() interface{} {
	if !v.CanAddr() {
		val := v.Value.Interface()
		fmt.Println(reflect.ValueOf(val).CanAddr(), reflect.ValueOf(val).Type(), reflect.ValueOf(val).IsNil())
		return reflect.ValueOf(&val).Interface()
	}
	return v.Addr().Interface()
}

func (v pointerValue) init() {
	strValue := reflect.New(v.Type().Elem()).Elem()
	v.Set(strValue.Addr())
}

func (v pointerValue) pointedValue() reflect.Value {
	return v.Elem()
}
