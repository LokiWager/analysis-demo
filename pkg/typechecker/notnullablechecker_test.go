package typechecker

import (
	"testing"
)

func TestChecker_NotNullable(t *testing.T) {
	err := RunChecker("Hello", "NotNullable", nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = RunChecker(nil, "NotNullable", nil)
	if err == nil {
		t.Errorf("Expected error for nil value, got none")
	}
}
