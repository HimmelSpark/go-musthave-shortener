package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/auth"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/stretchr/testify/require"
)

func newTestUserHandler() (*UserHandler, *mockShortenerService) {
	m := &mockShortenerService{}
	h := NewUserHandler(m)
	return h, m
}

func withAuth(req *http.Request, info auth.Info) *http.Request {
	return req.WithContext(auth.WithInfo(req.Context(), info))
}

func TestGetUserURLs_OK(t *testing.T) {
	h, mock := newTestUserHandler()
	mock.getUserURLsFn = func(userID string) ([]service.UserURLOutput, error) {
		require.Equal(t, "user-1", userID)
		return []service.UserURLOutput{
			{ShortURL: "http://localhost:8080/abc", OriginalURL: "https://yandex.ru"},
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = withAuth(req, auth.Info{UserID: "user-1", CookiePresent: true, CookieValid: true})
	rec := httptest.NewRecorder()

	h.GetUserURLs(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var got []map[string]string
	require.NoError(t, json.Unmarshal(body, &got))
	require.Len(t, got, 1)
	require.Equal(t, "http://localhost:8080/abc", got[0]["short_url"])
	require.Equal(t, "https://yandex.ru", got[0]["original_url"])
}

func TestGetUserURLs_NoContent(t *testing.T) {
	h, mock := newTestUserHandler()
	mock.getUserURLsFn = func(userID string) ([]service.UserURLOutput, error) {
		return nil, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = withAuth(req, auth.Info{UserID: "user-1", CookiePresent: false, CookieValid: false})
	rec := httptest.NewRecorder()

	h.GetUserURLs(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestGetUserURLs_Unauthorized(t *testing.T) {
	h, _ := newTestUserHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = withAuth(req, auth.Info{UserID: "user-1", CookiePresent: true, CookieValid: false})
	rec := httptest.NewRecorder()

	h.GetUserURLs(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}
