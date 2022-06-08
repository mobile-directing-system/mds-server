package controller

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test_generatePublicSessionToken tests generatePublicSessionToken.
func Test_generatePublicSessionToken(t *testing.T) {
	token, err := generatePublicSessionToken("meow", "ola")
	require.NoError(t, err, "should not fail")
	assert.NotEmpty(t, token, "token should not be empty")
}
