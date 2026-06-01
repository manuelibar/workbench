package errs

import (
	"errors"
	"maps"
	"slices"
)

type singleUnwrapper interface {
	Unwrap() error
}

type multiUnwrapper interface {
	Unwrap() []error
}

// CodeOf returns the first stable code found while traversing err.
func CodeOf(err error) Code {
	var code Code
	walk(err, func(cur error) bool {
		switch e := cur.(type) {
		case *Error:
			if e != nil && e.code != nil {
				code = *e.code
				return false
			}
		case *Multi:
			if e != nil && e.code != nil {
				code = *e.code
				return false
			}
		}
		return true
	})
	return code
}

// SentinelOf returns the first classification sentinel found while traversing err.
func SentinelOf(err error) error {
	var sentinel error
	walk(err, func(cur error) bool {
		switch e := cur.(type) {
		case *Error:
			if e != nil && e.sentinel != nil && *e.sentinel != nil {
				sentinel = *e.sentinel
				return false
			}
		case *Multi:
			if e != nil && e.sentinel != nil && *e.sentinel != nil {
				sentinel = *e.sentinel
				return false
			}
		}
		for _, known := range sentinels {
			if errors.Is(cur, known) {
				sentinel = known
				return false
			}
		}
		return true
	})
	return sentinel
}

// SeverityOf returns the first severity found while traversing err.
func SeverityOf(err error) Severity {
	var severity Severity
	walk(err, func(cur error) bool {
		switch e := cur.(type) {
		case *Error:
			if e != nil && e.severity != nil {
				severity = *e.severity
				return false
			}
		case *Multi:
			if e != nil && e.severity != nil {
				severity = *e.severity
				return false
			}
		}
		return true
	})
	return severity
}

// IsRetryable reports the first explicit retryable value found while traversing err.
func IsRetryable(err error) bool {
	var retryable bool
	walk(err, func(cur error) bool {
		switch e := cur.(type) {
		case *Error:
			if e != nil && e.retryable != nil {
				retryable = *e.retryable
				return false
			}
		case *Multi:
			if e != nil && e.retryable != nil {
				retryable = *e.retryable
				return false
			}
		}
		return true
	})
	return retryable
}

// AttrsOf returns merged attributes from err's tree.
//
// Attributes closer to the root override attributes with the same key deeper in
// the cause or child tree. The returned map is a shallow copy.
func AttrsOf(err error) map[string]any {
	attrs := make(map[string]any)
	attrsInto(err, attrs)
	if len(attrs) == 0 {
		return nil
	}
	return attrs
}

// StackOf returns the first captured stack found while traversing err.
func StackOf(err error) []Frame {
	var stack []Frame
	walk(err, func(cur error) bool {
		switch e := cur.(type) {
		case *Error:
			if e != nil && len(e.stack) > 0 {
				stack = slices.Clone(e.stack)
				return false
			}
		case *Multi:
			if e != nil && len(e.stack) > 0 {
				stack = slices.Clone(e.stack)
				return false
			}
		}
		return true
	})
	return stack
}

func walk(err error, visit func(error) bool) bool {
	if err == nil {
		return true
	}
	if !visit(err) {
		return false
	}
	if unwrapper, ok := err.(multiUnwrapper); ok {
		for _, child := range unwrapper.Unwrap() {
			if !walk(child, visit) {
				return false
			}
		}
		return true
	}
	if unwrapper, ok := err.(singleUnwrapper); ok {
		return walk(unwrapper.Unwrap(), visit)
	}
	return true
}

func attrsInto(err error, attrs map[string]any) {
	if err == nil {
		return
	}

	switch e := err.(type) {
	case *Error:
		attrsInto(e.cause, attrs)
		maps.Copy(attrs, e.attrs)
		return
	case *Multi:
		for _, child := range e.children {
			attrsInto(child, attrs)
		}
		maps.Copy(attrs, e.attrs)
		return
	}

	if unwrapper, ok := err.(multiUnwrapper); ok {
		for _, child := range unwrapper.Unwrap() {
			attrsInto(child, attrs)
		}
		return
	}
	if unwrapper, ok := err.(singleUnwrapper); ok {
		attrsInto(unwrapper.Unwrap(), attrs)
	}
}
