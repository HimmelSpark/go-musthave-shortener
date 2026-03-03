package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockShortenerService struct {
	shortenFn func(url string) (string, error)
	findFn    func(id string) (string, error)

	gotURL string
	gotID  string
}

func (m *mockShortenerService) ShortenURL(url string) (string, error) {
	m.gotURL = url
	if m.shortenFn != nil {
		return m.shortenFn(url)
	}
	return "", nil
}

func (m *mockShortenerService) FindURL(id string) (string, error) {
	m.gotID = id
	if m.findFn != nil {
		return m.findFn(id)
	}
	return "", nil
}

func newTestHandler() (*ShortenerHandler, *mockShortenerService) {
	m := &mockShortenerService{}
	h := NewShortenerHandler(m)
	return h, m
}

func TestCreateShortUrlOk(t *testing.T) {
	h, mock := newTestHandler()
	mock.shortenFn = func(url string) (string, error) {
		return "http://localhost:8080/AvAjEv0l", nil
	}

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader("https://yandex.ru"))
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()

	h.ShortenURL(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	body := string(b)

	require.Equal(t, http.StatusCreated, res.StatusCode)
	require.True(t, strings.HasPrefix(res.Header.Get("Content-Type"), "text/plain"))
	require.Equal(t, "http://localhost:8080/AvAjEv0l", body)

	require.Equal(t, "https://yandex.ru", mock.gotURL)
}

func TestGetRedirectURLOk(t *testing.T) {
	h, mock := newTestHandler()
	mock.findFn = func(id string) (string, error) {
		return "https://yandex.ru", nil
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/AvAjEv0l", nil)
	req.SetPathValue("urlId", "AvAjEv0l")
	rec := httptest.NewRecorder()

	h.GetRedirectURL(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	require.Equal(t, "https://yandex.ru", res.Header.Get("Location"))
	require.Equal(t, "AvAjEv0l", mock.gotID)
}
