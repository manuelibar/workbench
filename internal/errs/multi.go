package errs

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// Multi is a classified error for workflows that tolerate and aggregate
// multiple child failures.
type Multi struct {
	msg      string
	children []error
	metadata
}

// NewMulti creates a classified multi-error and captures the caller stack.
//
// Add initial children with WithCause. Nil children are ignored.
func NewMulti(msg string, opts ...Option) *Multi {
	cfg := options(opts)
	err := &Multi{
		msg: msg,
		metadata: metadata{
			stack: stack(3),
		},
	}
	err.Add(cfg.causes...)
	err.metadata.apply(cfg.meta)
	return err
}

// Add appends child errors. Nil children are ignored.
func (m *Multi) Add(children ...error) {
	if m == nil {
		return
	}
	for _, child := range children {
		if child != nil {
			m.children = append(m.children, child)
		}
	}
}

func (m *Multi) Len() int {
	if m == nil {
		return 0
	}
	return len(m.children)
}

func (m *Multi) Error() string {
	if m == nil {
		return "<nil>"
	}
	if m.msg != "" {
		if len(m.children) == 0 {
			return m.msg
		}
		return fmt.Sprintf("%s (%d errors)", m.msg, len(m.children))
	}
	switch len(m.children) {
	case 0:
		return ""
	case 1:
		return m.children[0].Error()
	default:
		return fmt.Sprintf("%d errors", len(m.children))
	}
}

// Unwrap returns the child errors. The returned slice is a copy.
func (m *Multi) Unwrap() []error {
	if m == nil || len(m.children) == 0 {
		return nil
	}
	return slices.Clone(m.children)
}

// Is reports whether target matches the stored sentinel.
func (m *Multi) Is(target error) bool {
	if m == nil || m.sentinel == nil || *m.sentinel == nil {
		return false
	}
	return errors.Is(*m.sentinel, target)
}

func (m *Multi) Message() string {
	if m == nil {
		return ""
	}
	return m.msg
}

func (m *Multi) Sentinel() error {
	if m == nil {
		return nil
	}
	if m.sentinel == nil {
		return nil
	}
	return *m.sentinel
}

func (m *Multi) Code() Code {
	if m == nil {
		return ""
	}
	if m.code == nil {
		return ""
	}
	return *m.code
}

func (m *Multi) Severity() Severity {
	if m == nil {
		return ""
	}
	if m.severity == nil {
		return ""
	}
	return *m.severity
}

func (m *Multi) Retryable() bool {
	if m == nil || m.retryable == nil {
		return false
	}
	return *m.retryable
}

// Attrs returns a shallow copy of the multi-error attributes.
func (m *Multi) Attrs() map[string]any {
	if m == nil || len(m.attrs) == 0 {
		return nil
	}
	return maps.Clone(m.attrs)
}

// Stack returns a copy of the captured stack.
func (m *Multi) Stack() []Frame {
	if m == nil || len(m.stack) == 0 {
		return nil
	}
	return slices.Clone(m.stack)
}

func (m *Multi) clone() *Multi {
	if m == nil {
		return nil
	}
	return &Multi{
		msg:      m.msg,
		children: slices.Clone(m.children),
		metadata: m.metadata.clone(),
	}
}
