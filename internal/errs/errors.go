package errs

import (
	"errors"
	"maps"
	"slices"
)

var (
	// ErrInvalid classifies input, state, or command errors caused by invalid data.
	ErrInvalid = errors.New("invalid")
	// ErrNotFound classifies failures where a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrDependencyFailed classifies failures caused by a required dependency.
	ErrDependencyFailed = errors.New("dependency failed")
	// ErrUnavailable classifies failures caused by temporary unavailability.
	ErrUnavailable = errors.New("unavailable")
	// ErrTimeout classifies failures caused by timeouts.
	ErrTimeout = errors.New("timeout")
)

var sentinels = []error{
	ErrInvalid,
	ErrNotFound,
	ErrDependencyFailed,
	ErrUnavailable,
	ErrTimeout,
}

// Code is a stable application-defined error code.
type Code string

func (c Code) String() string {
	return string(c)
}

// Severity describes the operational severity of an error.
type Severity string

const (
	// SeverityWarning classifies failures that should be noticed but are not urgent.
	SeverityWarning Severity = "warning"
	// SeverityError classifies ordinary errors.
	SeverityError Severity = "error"
	// SeverityCritical classifies urgent failures requiring immediate attention.
	SeverityCritical Severity = "critical"
)

// Frame identifies one captured call frame.
type Frame struct {
	Function string
	File     string
	Line     int
}

// Error is a classified error with one ordinary cause.
type Error struct {
	msg   string
	cause error
	metadata
}

// New creates a classified error and captures the caller stack.
func New(msg string, opts ...Option) *Error {
	cfg := options(opts)
	err := &Error{
		msg: msg,
		metadata: metadata{
			stack: stack(3),
		},
	}
	if cfg.cause != nil {
		err.cause = cfg.cause
	}
	err.metadata.apply(cfg.meta)
	return err
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.msg == "" {
		if e.cause != nil {
			return e.cause.Error()
		}
		if e.sentinel != nil && *e.sentinel != nil {
			return (*e.sentinel).Error()
		}
		return ""
	}
	if e.cause != nil {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

// Unwrap returns the ordinary cause. It does not return the sentinel.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Is reports whether target matches the stored sentinel.
func (e *Error) Is(target error) bool {
	if e == nil || e.sentinel == nil || *e.sentinel == nil {
		return false
	}
	return errors.Is(*e.sentinel, target)
}

func (e *Error) Message() string {
	if e == nil {
		return ""
	}
	return e.msg
}

func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Sentinel() error {
	if e == nil {
		return nil
	}
	if e.sentinel == nil {
		return nil
	}
	return *e.sentinel
}

func (e *Error) Code() Code {
	if e == nil {
		return ""
	}
	if e.code == nil {
		return ""
	}
	return *e.code
}

func (e *Error) Severity() Severity {
	if e == nil {
		return ""
	}
	if e.severity == nil {
		return ""
	}
	return *e.severity
}

func (e *Error) Retryable() bool {
	if e == nil || e.retryable == nil {
		return false
	}
	return *e.retryable
}

// Attrs returns a shallow copy of the error attributes.
func (e *Error) Attrs() map[string]any {
	if e == nil || len(e.attrs) == 0 {
		return nil
	}
	return maps.Clone(e.attrs)
}

// Stack returns a copy of the captured stack.
func (e *Error) Stack() []Frame {
	if e == nil || len(e.stack) == 0 {
		return nil
	}
	return slices.Clone(e.stack)
}

func (e *Error) clone() *Error {
	if e == nil {
		return nil
	}
	return &Error{
		msg:      e.msg,
		cause:    e.cause,
		metadata: e.metadata.clone(),
	}
}
