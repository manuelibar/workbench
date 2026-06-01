package errs

// Decorate returns err with patched metadata.
//
// Decorate means "same failure, more context". It never changes the cause:
// WithCause is ignored here. Other options patch metadata on the returned copy.
func Decorate(err error, opts ...Option) error {
	if err == nil {
		return nil
	}

	var cfg opt
	for _, opt := range opts {
		cfg.meta.apply(opt.meta)
	}

	switch e := err.(type) {
	case *Error:
		copied := e.clone()
		copied.metadata.apply(cfg.meta)
		return copied
	case *Multi:
		copied := e.clone()
		copied.metadata.apply(cfg.meta)
		return copied
	default:
		decorated := &Error{
			cause: err,
			metadata: metadata{
				stack: stack(3),
			},
		}
		decorated.metadata.apply(cfg.meta)
		return decorated
	}
}
