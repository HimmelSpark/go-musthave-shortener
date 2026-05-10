package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/auth"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/stretchr/testify/require"
)

type mockDeletionService struct {
	gotUserID string
	gotIDs    []string
	calls     int
}

func (m *mockDeletionService) Submit(userID string, shortIDs []string) {
	m.calls++
	m.gotUserID = userID
	m.gotIDs = append([]string(nil), shortIDs...)
}

func (m *mockDeletionService) Run(_ context.Context) {}

func newTestUserHandler() (*UserHandler, *mockShortenerService, *mockDeletionService) {
	m := &mockShortenerService{}
	d := &mockDeletionService{}
	h := NewUserHandler(m, d)
	return h, m, d
}

func withAuth(req *http.Request, info auth.Info) *http.Request {
	return req.WithContext(auth.WithInfo(req.Context(), info))
}

func TestGetUserURLs_OK(t *testing.T) {
	h, mock, _ := newTestUserHandler()
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
	h, mock, _ := newTestUserHandler()
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
	h, _, _ := newTestUserHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = withAuth(req, auth.Info{UserID: "user-1", CookiePresent: true, CookieValid: false})
	rec := httptest.NewRecorder()

	h.GetUserURLs(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestDeleteUserURLs_Accepted(t *testing.T) {
	h, _, del := newTestUserHandler()

	body := strings.NewReader(`["abc","def","ghi"]`)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	req = withAuth(req, auth.Info{UserID: "user-1", CookiePresent: true, CookieValid: true})
	rec := httptest.NewRecorder()

	h.DeleteUserURLs(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusAccepted, res.StatusCode)
	require.Equal(t, 1, del.calls)
	require.Equal(t, "user-1", del.gotUserID)
	require.Equal(t, []string{"abc", "def", "ghi"}, del.gotIDs)
}

func TestDeleteUserURLs_Unauthorized(t *testing.T) {
	h, _, del := newTestUserHandler()

	body := strings.NewReader(`["abc"]`)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	req = withAuth(req, auth.Info{UserID: "user-1", CookiePresent: true, CookieValid: false})
	rec := httptest.NewRecorder()

	h.DeleteUserURLs(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	require.Equal(t, 0, del.calls)
}

