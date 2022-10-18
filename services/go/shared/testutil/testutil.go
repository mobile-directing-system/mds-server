// Package testutil provides different utils for testing.
package testutil

import (
	"github.com/gofrs/uuid"
	"math/rand"
	"time"
)

// NewUUIDV4 returns a new v4 uuid.UUID. If generation fails, it panicks.
func NewUUIDV4() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id
}

// NewRandomTime creates a new time.Time with random value.
func NewRandomTime() time.Time {
	t := time.UnixMicro(rand.Int63())
	t = t.In(time.UTC)
	// Limit year because of JSON marshalling.
	if t.Year() >= 9999 {
		t = t.AddDate(-t.Year()+9900, 0, 0)
	}
	return t
}
