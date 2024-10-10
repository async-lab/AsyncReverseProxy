package util

import "reflect"

func GetStructValue(v interface{}) reflect.Value {
	if v == nil {
		return reflect.Value{}
	}
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func GetStructType(v interface{}) reflect.Type {
	tType := reflect.TypeOf(v)
	for tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	return tType
}
