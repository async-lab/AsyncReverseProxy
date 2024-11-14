package lang

import (
	"reflect"
)

func GetForActualValue(v interface{}) reflect.Value {
	if v == nil {
		return reflect.Value{}
	}
	val := reflect.ValueOf(v)
	for {
		switch val.Kind() {
		case reflect.Ptr:
			val = val.Elem()
		case reflect.Interface:
			val = val.Elem()
		default:
			return val
		}
	}
}

func GetForActualType(v interface{}) reflect.Type {
	tType := reflect.TypeOf(v)
	for {
		switch tType.Kind() {
		case reflect.Ptr:
			tType = tType.Elem()
		case reflect.Interface:
			tType = tType.Elem()
		default:
			return tType
		}
	}
}

func GetForPtrType(v interface{}) reflect.Type {
	return reflect.PtrTo(GetForActualType(v))
}

func GetForActualTypeWithType[T any]() reflect.Type {
	return GetForActualType((*T)(nil))
}

func GetForPtrTypeWithType[T any]() reflect.Type {
	return reflect.PtrTo(GetForActualTypeWithType[T]())
}
