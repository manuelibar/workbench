package domain

import "github.com/google/uuid"

// User is the bootstrap user record. v0 has exactly one row in the users
// table per workbench instance; multi-user is post-v0.
type User struct {
	ID          uuid.UUID
	DisplayName string
}
