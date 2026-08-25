package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("hunter2")
	require.NoError(t, err)
	require.NotEqual(t, "hunter2", hash) // stored value is a hash, not the password

	require.True(t, CheckPassword("hunter2", hash))
	require.False(t, CheckPassword("wrong", hash))

	require.NoError(t, CompareHashAndPassword(hash, "hunter2"))
	require.Error(t, CompareHashAndPassword(hash, "wrong"))
}

func TestHashPassword_SaltedPerCall(t *testing.T) {
	h1, err := HashPassword("same")
	require.NoError(t, err)
	h2, err := HashPassword("same")
	require.NoError(t, err)
	require.NotEqual(t, h1, h2) // bcrypt embeds a random salt
}
