package id

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// New returns a fresh UUIDv7 (time-ordered, suitable for primary keys).
//
// New panics on the impossible case where crypto/rand fails — UUIDv7
// generation has no other failure mode.
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("id: NewV7: %v", err))
	}
	return id
}

// Audit is the four-part identifier set carried alongside every
// state-bearing row and propagated through every tool invocation.
//
//   - RequestID is unique to one inbound request.
//   - CorrelationID groups all causally-related work for a higher-level
//     operation (e.g. a Run).
//   - CausationID points at the request that directly caused this one.
//   - IdempotencyKey is supplied by the caller to deduplicate retries.
//
// Zero values are valid; pgstore writes them as SQL NULL.
type Audit struct {
	RequestID      uuid.UUID
	CorrelationID  uuid.UUID
	CausationID    uuid.UUID
	IdempotencyKey string
}

type auditCtxKey struct{}

// WithAudit returns a copy of ctx carrying the supplied [Audit] values.
func WithAudit(ctx context.Context, a Audit) context.Context {
	return context.WithValue(ctx, auditCtxKey{}, a)
}

// FromContext returns the [Audit] previously attached with [WithAudit], or
// the zero value (with ok == false) if none is present.
func FromContext(ctx context.Context) (Audit, bool) {
	a, ok := ctx.Value(auditCtxKey{}).(Audit)
	return a, ok
}

// EnsureRequest returns ctx with an [Audit] guaranteed to have a non-zero
// RequestID. If ctx already carries one, it is returned unchanged; otherwise
// a fresh UUIDv7 is generated and attached.
func EnsureRequest(ctx context.Context) (context.Context, Audit) {
	a, ok := FromContext(ctx)
	if ok && a.RequestID != uuid.Nil {
		return ctx, a
	}
	a.RequestID = New()
	return WithAudit(ctx, a), a
}
