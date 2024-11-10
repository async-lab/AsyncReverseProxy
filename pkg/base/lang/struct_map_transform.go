package lang

import "reflect"

func StructToMap[T any](v T) map[string]interface{} {
	result := make(map[string]interface{})
	val := GetForStructValue(v)

	for i := 0; i < val.Type().NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		fieldValue := val.Field(i).Interface()
		result[fieldName] = fieldValue
	}

	return result
}

func MapToStruct[T any](m map[string]interface{}, v T) T {
	val := GetForStructValue(v)

	for i := 0; i < val.Type().NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		fieldValue := m[fieldName]
		val.Field(i).Set(reflect.ValueOf(fieldValue))
	}

	return v
}
