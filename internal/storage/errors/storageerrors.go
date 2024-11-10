package storerrors

import "errors"

var (
	ErrorUserNotFound    = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserExists        = errors.New("user already exists")
	ErrUserDoesNotExists = errors.New("user does not exists")
	ErrBookDoesNotExists = errors.New("book does not exists")
	ErrEmptyBookList     = errors.New("empty book list")
)
