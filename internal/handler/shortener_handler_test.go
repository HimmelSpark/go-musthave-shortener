package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

type mockShortenerService struct {
	shortenFn func(url, baseURL string) (string, error)
	findFn    func(id string) (string, error)

	gotURL     string
	gotBaseURL string
	gotID      string
}

func (m *mockShortenerService) ShortenURL(url, baseURL string) (string, error) {
	m.gotURL = url
	m.gotBaseURL = baseURL
	if m.shortenFn != nil {
		return m.shortenFn(url, baseURL)
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
	mock.shortenFn = func(url, baseURL string) (string, error) {
		return "AvAjEv0l", nil
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
	require.NotEmpty(t, body)
	require.Equal(t, 8, utf8.RuneCountInString(body))

	require.Equal(t, "https://yandex.ru", mock.gotURL)
	require.Equal(t, "http://localhost:8080", mock.gotBaseURL)
}

/*
Можно еще много кейсов накидать, но пока лень 🦥
*/
