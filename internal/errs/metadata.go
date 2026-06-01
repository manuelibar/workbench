package errs

import (
	"maps"
	"slices"

	"github.com/manuelibar/workbench/internal/ptr"
)

type metadata struct {
	sentinel  *error
	code      *Code
	severity  *Severity
	retryable *bool
	attrs     map[string]any
	stack     []Frame
}

func (m metadata) clone() metadata {
	if m.sentinel != nil {
		m.sentinel = ptr.Ptr(*m.sentinel)
	}
	if m.code != nil {
		m.code = ptr.Ptr(*m.code)
	}
	if m.severity != nil {
		m.severity = ptr.Ptr(*m.severity)
	}
	if m.retryable != nil {
		m.retryable = ptr.Ptr(*m.retryable)
	}
	if len(m.attrs) > 0 {
		m.attrs = maps.Clone(m.attrs)
	} else {
		m.attrs = nil
	}
	if len(m.stack) > 0 {
		m.stack = slices.Clone(m.stack)
	} else {
		m.stack = nil
	}
	return m
}

func (m *metadata) apply(p metadata) {
	if p.sentinel != nil {
		m.sentinel = ptr.Ptr(*p.sentinel)
	}
	if p.code != nil {
		m.code = ptr.Ptr(*p.code)
	}
	if p.severity != nil {
		m.severity = ptr.Ptr(*p.severity)
	}
	if p.retryable != nil {
		m.retryable = ptr.Ptr(*p.retryable)
	}
	if len(p.attrs) > 0 {
		if m.attrs == nil {
			m.attrs = make(map[string]any, len(p.attrs))
		}
		maps.Copy(m.attrs, p.attrs)
	}
}
