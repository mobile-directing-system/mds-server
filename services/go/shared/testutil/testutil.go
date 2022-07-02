// Package testutil provides different utils for testing.
package testutil

import "github.com/gofrs/uuid"

// NewUUIDV4 returns a new v4 uuid.UUID. If generation fails, it panicks.
func NewUUIDV4() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id
}
