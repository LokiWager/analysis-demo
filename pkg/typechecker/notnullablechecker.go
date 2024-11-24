package typechecker

import (
	"errors"
	"log"
	"reflect"
)

// NotNullableChecker is a type that implements the TypeChecker interface.
type NotNullableChecker struct{}

func NewNotNullableChecker() *NotNullableChecker {
	return &NotNullableChecker{}
}

func (n *NotNullableChecker) Check(value interface{}) error {
	if value == nil {
		return errors.New("value is nil")
	}

	v := reflect.ValueOf(value)

	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return errors.New("value is nil (pointer or interface)")
	}

	if v.Kind() == reflect.Slice && v.Len() == 0 {
		return errors.New("value is an empty slice")
	}

	if v.Kind() == reflect.Map && v.Len() == 0 {
		return errors.New("value is an empty map")
	}

	if v.Kind() == reflect.String && v.Len() == 0 {
		return errors.New("value is an empty string")
	}

	if v.Kind() == reflect.Array && v.Len() == 0 {
		return errors.New("value is an empty array")
	}

	if v.Kind() == reflect.Chan && v.Len() == 0 {
		return errors.New("value is an empty channel")
	}

	if v.Kind() == reflect.Func && v.IsNil() {
		return errors.New("value is a nil function")
	}

	if v.Kind() == reflect.Struct && v.NumField() == 0 {
		return errors.New("value is an empty struct")
	}

	return nil
}

func NotNullable(value interface{}) interface{} {
	checker := NewNotNullableChecker()
	err := checker.Check(value)
	if err != nil {
		log.Printf("NotNullable check failed: %v", err)
		return nil
	}
	return value
}

func init() {
	RegisterChecker("NotNullable", func(params []string) TypeChecker {
		return NewNotNullableChecker()
	})
}
