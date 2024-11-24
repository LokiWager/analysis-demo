package typechecker

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCheckerAnalyzer(t *testing.T) {
	testData := analysistest.TestData()
	analysistest.Run(t, testData, CheckerAnalyzer, "example")
}

func TestChecker_Invalid(t *testing.T) {
	err := RunChecker(50, "NonExistentChecker", nil)
	if err == nil {
		t.Errorf("Expected error for non-existent checker, got none")
	}
}

func TestAnnotation_ParseComment(t *testing.T) {
	tests := []struct {
		comment      string
		expectedName string
		expectedArgs []string
		expectError  bool
	}{
		{"@check:NotNullable", "NotNullable", nil, false},
		{"@check:Range:10,100", "Range", []string{"10", "100"}, false},
		{"@check:InvalidChecker", "InvalidChecker", nil, false},
		{"InvalidFormat", "", nil, true},
	}

	for _, test := range tests {
		name, args, err := ParseComment(test.comment)
		if (err != nil) != test.expectError {
			t.Errorf("Unexpected error result for %s: got %v, want error=%v", test.comment, err, test.expectError)
		}
		if name != test.expectedName {
			t.Errorf("Unexpected name for %s: got %s, want %s", test.comment, name, test.expectedName)
		}
		if len(args) != len(test.expectedArgs) {
			t.Errorf("Unexpected argument count for %s: got %v, want %v", test.comment, args, test.expectedArgs)
		}
		for i := range args {
			if args[i] != test.expectedArgs[i] {
				t.Errorf("Unexpected argument for %s: got %v, want %v", test.comment, args, test.expectedArgs)
			}
		}
	}
}

func TestAnnotation_RunChecker(t *testing.T) {
	RegisterChecker("Range", func(params []string) TypeChecker {
		if len(params) != 2 {
			t.Fatalf("Range checker requires 2 parameters: min and max")
		}
		minN := 10.0
		maxN := 100.0
		return NewRangeChecker(minN, maxN)
	})

	tests := []struct {
		comment     string
		value       interface{}
		expectError bool
	}{
		{"@check:NotNullable", "Hello", false},
		{"@check:NotNullable", nil, true},
		{"@check:Range:10,100", 50, false},
		{"@check:Range:10,100", 200, true},
		{"@check:Range:10,100", "Invalid", true},
	}

	for _, test := range tests {
		name, args, err := ParseComment(test.comment)
		if err != nil {
			t.Errorf("Failed to parse comment %s: %v", test.comment, err)
			continue
		}
		err = RunChecker(test.value, name, args)
		if (err != nil) != test.expectError {
			t.Errorf("Unexpected checker result for %s (value: %v): got %v, want error=%v", test.comment, test.value, err, test.expectError)
		}
	}
}

func TestAnnotation_UnregisteredChecker(t *testing.T) {
	comment := "@check:NonExistentChecker"
	value := 42
	name, args, err := ParseComment(comment)
	if err != nil {
		t.Fatalf("Failed to parse comment: %v", err)
	}

	err = RunChecker(value, name, args)
	if err == nil {
		t.Errorf("Expected error for unregistered checker, got none")
	}
}
