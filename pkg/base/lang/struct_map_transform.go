package lang

import (
	"fmt"
	"reflect"
)

func StructToMap(v any) map[string]interface{} {
	result := make(map[string]interface{})
	val := GetForActualValue(v)

	for i := 0; i < val.Type().NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		fieldValue := val.Field(i).Interface()
		result[fieldName] = fieldValue
	}

	return result
}

func MapToStruct[T any](m map[string]interface{}, v *T) error {
	val := GetForActualValue(v)

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
