package lang

import (
	"reflect"
)

func GetActualValue(v interface{}) reflect.Value {
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

func GetActualType(v interface{}) reflect.Type {
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

func GetActualPtrType(v interface{}) reflect.Type {
	return reflect.PtrTo(GetActualType(v))
}

func GetActualTypeWithGeneric[T any]() reflect.Type {
	return GetActualType((*T)(nil))
}

func GetActualPtrTypeWithGeneric[T any]() reflect.Type {
	return reflect.PtrTo(GetActualTypeWithGeneric[T]())
}

type FQN string

func (fqn FQN) String() string {
	return string(fqn)
}

func GetFQN(t reflect.Type) FQN {
	return FQN(t.PkgPath() + "." + t.Name())
}
