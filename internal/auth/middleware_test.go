package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareNoCookieIssuesNew(t *testing.T) {
	const secret = "secret"
	var captured Info

	h := Middleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.NotEmpty(t, captured.UserID)
	require.False(t, captured.CookiePresent)
	require.False(t, captured.CookieValid)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, CookieName, cookies[0].Name)
}

func TestMiddlewareValidCookiePassesThrough(t *testing.T) {
	const secret = "secret"
	const userID = "user-abc"
	var captured Info

	h := Middleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: userID + "|" + Sign(userID, secret)})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, userID, captured.UserID)
	require.True(t, captured.CookiePresent)
	require.True(t, captured.CookieValid)

	require.Empty(t, rec.Result().Cookies())
}

func TestMiddlewareTamperedCookieReissuedAndFlagsSet(t *testing.T) {
	const secret = "secret"
	var captured Info

	h := Middleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "user-abc|deadbeef"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.NotEmpty(t, captured.UserID)
	require.NotEqual(t, "user-abc", captured.UserID)
	require.True(t, captured.CookiePresent)
	require.False(t, captured.CookieValid)

	require.Len(t, rec.Result().Cookies(), 1)
}

func TestMiddlewareMalformedCookieReissued(t *testing.T) {
	const secret = "secret"
	var captured Info

	h := Middleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "no-separator"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.NotEmpty(t, captured.UserID)
	require.True(t, captured.CookiePresent)
	require.False(t, captured.CookieValid)
}
