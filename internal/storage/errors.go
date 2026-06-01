package storage

import (
	"errors"
	"fmt"
)

var (
	ErrInvalid          = errors.New("storage: invalid")
	ErrNotFound         = errors.New("storage: not found")
	ErrDependencyFailed = errors.New("storage: dependency failed")
)

type Error struct {
	Kind    error
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return "storage error"
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	return target != nil && e != nil && errors.Is(e.Kind, target)
}

func invalid(code, format string, args ...any) error {
	return &Error{Kind: ErrInvalid, Code: code, Message: fmt.Sprintf(format, args...)}
}

func notFound(code, format string, args ...any) error {
	return &Error{Kind: ErrNotFound, Code: code, Message: fmt.Sprintf(format, args...)}
}

func dependency(code, message string, err error) error {
	return &Error{Kind: ErrDependencyFailed, Code: code, Message: message, Err: err}
}
