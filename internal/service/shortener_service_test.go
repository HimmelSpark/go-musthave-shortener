package service

import (
	"testing"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestShortenerService_FindURL_OK(t *testing.T) {
	repo := repository.NewInMemoryURLRepository()
	svc := NewShortenerService(repo)

	ok, err := repo.CreateURL("https://yandex.ru", "AvAjEv0l")
	require.NoError(t, err)
	require.True(t, ok)

	got, err := svc.FindURL("AvAjEv0l")
	require.NoError(t, err)
	require.Equal(t, "https://yandex.ru", got)
}

func TestShortenerService_ShortenURL_OK(t *testing.T) {
	repo := repository.NewInMemoryURLRepository()
	svc := NewShortenerService(repo)

	host := "http://localhost:8080"
	orig := "  https://yandex.ru  "

	short, err := svc.ShortenURL(orig, host)
	require.NoError(t, err)

	require.Contains(t, short, host+"/")

	shortID := short[len(host)+1:]
	got, err := svc.FindURL(shortID)
	require.NoError(t, err)
	require.Equal(t, "https://yandex.ru", got)
}
