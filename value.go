package sqltmpl

import (
	"database/sql/driver"
	"reflect"
)

func getFieldValue(rv reflect.Value, fieldName string) reflect.Value {
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if kind := rv.Kind(); kind == reflect.Map {
		return rv.MapIndex(reflect.ValueOf(fieldName))
	} else if kind == reflect.Struct {
		return rv.FieldByName(fieldName)
	}
	return _EmptyValue
}

func getValue(dotValue reflect.Value, path ...string) reflect.Value {
	value := dotValue
	for _, field := range path {
		value = getFieldValue(value, field)
	}
	return value
}

func toRawValue(rv reflect.Value) (value any, err error) {
	if !rv.IsValid() {
		// TODO: Return error?
		return
	}
	value = rv.Interface()
	if dv, ok := value.(driver.Valuer); ok {
		value, err = dv.Value()
	}
	return
}

func toBoolValue(rv reflect.Value) bool {
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() != 0
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() != 0
	case reflect.Interface, reflect.Pointer:
		return !rv.IsNil()
	}
	return false
}
