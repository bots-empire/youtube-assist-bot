package model

const (
	// ErrCommandNotConverted error command not converted.
	ErrCommandNotConverted = Error("command not converted")
	// ErrTaskNotFound error task not found.
	ErrTaskNotFound = Error("task not found")
	// ErrUserNotFound error user not found.
	ErrUserNotFound = Error("user not found")
	// ErrFoundTwoUsers error found two user account for one user.
	ErrFoundTwoUsers = Error("found two users")

	// ErrScanSqlRow error scan sql row.
	ErrScanSqlRow = Error("failed scan sql row")
)

type Error string

func (e Error) Error() string {
	return string(e)
}
