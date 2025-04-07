package lang

import (
	"fmt"
	"reflect"
)

func StructToMap(v any) map[string]interface{} {
	result := make(map[string]interface{})
	val := GetActualValue(v)

	if val.Kind() != reflect.Struct {
		panic("v must be a struct")
	}

	for i := 0; i < val.Type().NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		fieldValue := val.Field(i)

		if fieldValue.Kind() == reflect.Struct {
			result[fieldName] = StructToMap(fieldValue.Interface())
		} else {
			result[fieldName] = fieldValue.Interface()
		}

		result[fieldName] = fieldValue
	}

	return result
}

// TODO 没写递归
func MapToStruct[T any](m map[string]interface{}, v *T) error {
	val := GetActualValue(v)

	if val.Kind() != reflect.Struct {
		panic("v must be a struct pointer")
	}

	for i := 0; i < val.Type().NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		fieldValue, ok := m[fieldName]
		if !ok {
			return fmt.Errorf("field %s not found", fieldName)
		}
		val.Field(i).Set(reflect.ValueOf(fieldValue))
	}

	return nil
}
