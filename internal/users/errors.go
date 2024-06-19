package users

import "errors"

var (
	ErrInsertUser       = errors.New("db insert user error: ")
	ErrUserNotFound     = errors.New("db user not found")
	ErrPasswordMismatch = errors.New("db user password mismatch")
)
