package typechecker

import "testing"

func TestChecker_Range(t *testing.T) {
	RegisterChecker("Range", func(params []string) TypeChecker {
		minN := 10.0
		maxN := 100.0
		return NewRangeChecker(minN, maxN)
	})

	err := RunChecker(50, "Range", []string{"10", "100"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = RunChecker(200, "Range", []string{"10", "100"})
	if err == nil {
		t.Errorf("Expected error for out-of-range value, got none")
	}

	err = RunChecker(10, "Range", []string{"10", "100"})
	if err != nil {
		t.Errorf("Expected no error for lower boundary, got: %v", err)
	}

	err = RunChecker(100, "Range", []string{"10", "100"})
	if err != nil {
		t.Errorf("Expected no error for upper boundary, got: %v", err)
	}

}
