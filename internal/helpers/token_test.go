package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthToken_RoundTrip(t *testing.T) {
	secret := []byte("test-secret")
	tok, err := CreateAuthToken(42, secret)
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	uid, err := ParseAuthToken(tok, secret)
	require.NoError(t, err)
	require.Equal(t, int64(42), uid)
}

func TestParseAuthToken_WrongSecret(t *testing.T) {
	tok, err := CreateAuthToken(42, []byte("right-secret"))
	require.NoError(t, err)

	_, err = ParseAuthToken(tok, []byte("wrong-secret"))
	require.Error(t, err) // signature must not verify
}

func TestParseAuthToken_Garbage(t *testing.T) {
	_, err := ParseAuthToken("not-a-valid-token", []byte("secret"))
	require.Error(t, err)
}

func TestUniqueLogin(t *testing.T) {
	a := UniqueLogin("prefix")
	b := UniqueLogin("prefix")
	require.NotEqual(t, a, b)        // unique each call
	require.Contains(t, a, "prefix") // keeps the prefix
}
