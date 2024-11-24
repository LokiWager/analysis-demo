package typechecker

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// RangeChecker is a type that checks if a value is within a specified range.
type RangeChecker struct {
	Min, Max float64
}

func NewRangeChecker(min, max float64) *RangeChecker {
	return &RangeChecker{Min: min, Max: max}
}

func (r *RangeChecker) Check(value interface{}) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return errors.New("value is invalid")
	}

	if v.Kind() < reflect.Int || v.Kind() > reflect.Float64 {
		return errors.New("value is not a number")
	}

	num := v.Convert(reflect.TypeOf(float64(0))).Float()
	if num < r.Min || num > r.Max {
		return fmt.Errorf("value %v is out of range [%v, %v]", num, r.Min, r.Max)
	}

	return nil
}

func Range(value interface{}, min, max float64) interface{} {
	checker := NewRangeChecker(min, max)
	err := checker.Check(value)
	if err != nil {
		log.Printf("Range check failed: %v", err)
		return nil
	}
	return value
}

func init() {
	RegisterChecker("Range", func(params []string) TypeChecker {
		if len(params) != 2 {
			log.Fatalf("Range checker requires 2 parameters: min and max")
		}
		minN, err1 := strconv.ParseFloat(params[0], 64)
		maxN, err2 := strconv.ParseFloat(params[1], 64)
		if err1 != nil || err2 != nil {
			log.Fatalf("Invalid range parameters: %v, %v", params[0], params[1])
		}
		return NewRangeChecker(minN, maxN)
	})
}
