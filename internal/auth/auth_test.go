package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	sig := Sign("user-1", "secret")
	require.True(t, Verify("user-1", sig, "secret"))
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	sig := Sign("user-1", "secret")
	require.False(t, Verify("user-1", sig, "other-secret"))
}

func TestVerifyRejectsWrongPayload(t *testing.T) {
	sig := Sign("user-1", "secret")
	require.False(t, Verify("user-2", sig, "secret"))
}

func TestVerifyRejectsBadHex(t *testing.T) {
	require.False(t, Verify("user-1", "not-hex", "secret"))
}

func TestParseCookieValueOK(t *testing.T) {
	userID := "user-1"
	value := userID + "|" + Sign(userID, "secret")
	got, ok := ParseCookieValue(value, "secret")
	require.True(t, ok)
	require.Equal(t, userID, got)
}

func TestParseCookieValueMalformed(t *testing.T) {
	_, ok := ParseCookieValue("nopipehere", "secret")
	require.False(t, ok)
}

func TestParseCookieValueTampered(t *testing.T) {
	_, ok := ParseCookieValue("user-1|deadbeef", "secret")
	require.False(t, ok)
}
