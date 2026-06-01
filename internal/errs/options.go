package errs

import (
	"maps"

	"github.com/manuelibar/workbench/internal/ptr"
)

// Option configures errors created by New, NewMulti, or Decorate.
// Decorate ignores WithCause.
type Option struct {
	cause error
	meta  metadata
}

type opt struct {
	cause  error
	causes []error
	meta   metadata
}

// WithCause stores err as the ordinary cause. For NewMulti, WithCause appends
// err as a child. For Decorate, WithCause is ignored because decoration keeps
// the same failure cause. Nil causes are ignored.
func WithCause(err error) Option {
	return Option{cause: err}
}

// WithSentinel stores sentinel as the classification used by errors.Is.
func WithSentinel(sentinel error) Option {
	return Option{meta: metadata{sentinel: ptr.Ptr(sentinel)}}
}

// WithCode stores a stable application-defined code.
func WithCode(code Code) Option {
	return Option{meta: metadata{code: ptr.Ptr(code)}}
}

func WithSeverity(severity Severity) Option {
	return Option{meta: metadata{severity: ptr.Ptr(severity)}}
}

// WithRetryable stores whether retrying the failed operation is appropriate.
func WithRetryable(retryable bool) Option {
	return Option{meta: metadata{retryable: ptr.Ptr(retryable)}}
}

// WithAttr stores one contextual attribute. Attribute values are copied by
// reference, not deep-copied.
func WithAttr(key string, value any) Option {
	return Option{meta: metadata{attrs: map[string]any{key: value}}}
}

// WithAttrs stores contextual attributes. The map is shallow-copied when the
// option is created and again when attributes are returned.
func WithAttrs(attrs map[string]any) Option {
	return Option{meta: metadata{attrs: maps.Clone(attrs)}}
}

func options(opts []Option) opt {
	var cfg opt
	for _, opt := range opts {
		opt.applyTo(&cfg)
	}
	return cfg
}

func (o Option) applyTo(cfg *opt) {
	if o.cause != nil {
		cfg.cause = o.cause
		cfg.causes = append(cfg.causes, o.cause)
	}
	cfg.meta.apply(o.meta)
}
