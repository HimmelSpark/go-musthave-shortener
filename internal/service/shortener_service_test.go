package service

import (
	"testing"

	config "github.com/HimmelSpark/go-musthave-shortener.git/internal/config/shortener"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestShortenerService_FindURL_OK(t *testing.T) {
	repo := repository.NewInMemoryURLRepository()
	baseURL := "http://localhost:8080"
	svc, err := NewShortenerService(repo, &config.ServiceConfig{BaseURL: &baseURL})
	require.NoError(t, err)

	ok, err := repo.CreateURL("https://yandex.ru", "AvAjEv0l", "")
	require.NoError(t, err)
	require.True(t, ok)

	got, err := svc.FindURL("AvAjEv0l")
	require.NoError(t, err)
	require.Equal(t, "https://yandex.ru", got)
}

func TestShortenerService_ShortenURL_OK(t *testing.T) {
	repo := repository.NewInMemoryURLRepository()
	host := "http://localhost:8080"
	svc, err := NewShortenerService(repo, &config.ServiceConfig{BaseURL: &host})
	require.NoError(t, err)

	orig := "  https://yandex.ru  "

	short, err := svc.ShortenURL(orig, "user-1")
	require.NoError(t, err)

	require.Contains(t, short, host+"/")

	shortID := short[len(host)+1:]
	got, err := svc.FindURL(shortID)
	require.NoError(t, err)
	require.Equal(t, "https://yandex.ru", got)
}

func TestShortenerService_GetUserURLs_IsolatedPerUser(t *testing.T) {
	repo := repository.NewInMemoryURLRepository()
	host := "http://localhost:8080"
	svc, err := NewShortenerService(repo, &config.ServiceConfig{BaseURL: &host})
	require.NoError(t, err)

	_, err = svc.ShortenURL("https://yandex.ru", "userA")
	require.NoError(t, err)
	_, err = svc.ShortenURL("https://google.com", "userA")
	require.NoError(t, err)
	_, err = svc.ShortenURL("https://example.com", "userB")
	require.NoError(t, err)

	a, err := svc.GetUserURLs("userA")
	require.NoError(t, err)
	require.Len(t, a, 2)

	b, err := svc.GetUserURLs("userB")
	require.NoError(t, err)
	require.Len(t, b, 1)
	require.Equal(t, "https://example.com", b[0].OriginalURL)

	none, err := svc.GetUserURLs("userC")
	require.NoError(t, err)
	require.Empty(t, none)
}
