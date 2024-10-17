package util

import (
	"reflect"

	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

func GetForStructValue(v interface{}) reflect.Value {
	if v == nil {
		return reflect.Value{}
	}
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func GetForStructType(v interface{}) reflect.Type {
	tType := reflect.TypeOf(v)
	for tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	return tType
}

func GetForPtrType(v interface{}) reflect.Type {
	return reflect.PtrTo(GetForStructType(v))
}

func GetForStructTypeWithType[T any]() reflect.Type {
	return GetForStructType((*T)(nil))
}

func GetForPtrTypeWithType[T any]() reflect.Type {
	return reflect.PtrTo(GetForStructTypeWithType[T]())
}
