package typechecker

import (
	"errors"
	"fmt"
	"strings"
)

var CheckerRegistry = map[string]TypeCheckerFactory{}

type TypeCheckerFactory func(params []string) TypeChecker

func RegisterChecker(name string, factory TypeCheckerFactory) {
	CheckerRegistry[name] = factory
}

func ParseComment(comment string) (name string, params []string, err error) {
	if !strings.HasPrefix(comment, "@check:") {
		return "", nil, errors.New("not a valid checker comment")
	}

	parts := strings.Split(strings.TrimPrefix(comment, "@check:"), ":")
	name = parts[0]
	if len(parts) > 1 {
		params = strings.Split(parts[1], ",")
	}
	return name, params, nil
}

func RunChecker(value interface{}, name string, params []string) error {
	factory, exists := CheckerRegistry[name]
	if !exists {
		return fmt.Errorf("checker %s not found", name)
	}
	checker := factory(params)
	return checker.Check(value)
}
