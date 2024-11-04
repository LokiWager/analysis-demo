package rest

import (
	"encoding/json"
	"fmt"

	echo "github.com/labstack/echo/v4"
)

type (
	// Err is the error format for all apis.
	Err struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// NewErr creates an error.
func NewErr(code int, message string) *Err {
	return &Err{
		Code:    code,
		Message: message,
	}
}

func (e *Err) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func handleErr(ctx echo.Context, err *Err) {
	buff, err1 := json.Marshal(err)
	if err1 != nil {
		panic(err1)
	}
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(err.Code)
	_, _ = ctx.Response().Write(buff)
}
