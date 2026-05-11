package backlogclient

import "errors"

// Sentinel errors returned by [Client] methods. Match with [errors.Is]. The
// underlying server-side error message is wrapped with %w.
var (
	// ErrNotFound corresponds to HTTP 404 (issue not present).
	ErrNotFound = errors.New("backlogclient: not found")
	// ErrVersionConflict corresponds to HTTP 409 (OCC version mismatch).
	ErrVersionConflict = errors.New("backlogclient: version conflict")
	// ErrValidation corresponds to HTTP 400 (malformed body / invalid fields).
	ErrValidation = errors.New("backlogclient: validation error")
	// ErrServer covers every 5xx (and 4xx outside the sentinel codes).
	ErrServer = errors.New("backlogclient: server error")
)
