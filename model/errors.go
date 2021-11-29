package model

const (
	// ErrCommandNotConverted error command not converted.
	ErrCommandNotConverted = Error("command not converted")
	// ErrTaskNotFound error task not found.
	ErrTaskNotFound = Error("task not found")
)

type Error string

func (e Error) Error() string {
	return string(e)
}
