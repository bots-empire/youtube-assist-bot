package err

import "fmt"

type Error struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("servises err: code=%d name=%s", e.Code, e.Name)
}

var (
	ErrCommandNotConverted = &Error{
		Code: 1,
		Name: "ERROR_COMMAND_NOT_CONVERTED",
	}
	ErrTaskNotFound = &Error{
		Code: 2,
		Name: "ERROR_TASK_NOT_FOUND",
	}
)
