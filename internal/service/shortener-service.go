package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type ShortenerService interface {
	ShortenURL(url string, hostURL string) (string, error)
	FindURL(shortID string) (string, error)
}
type shortenerService struct {
	urlRepo repository.URLRepository
}

func NewShortenerService(urlRepo repository.URLRepository) ShortenerService {
	return &shortenerService{urlRepo: urlRepo}
}

func (s *shortenerService) FindURL(shortID string) (string, error) {
	shortID = strings.TrimSpace(shortID)
	url, err := s.urlRepo.FindURL(shortID)
	return url, err
}

func (s *shortenerService) ShortenURL(url string, hostURL string) (string, error) {
	url = strings.TrimSpace(url)
	for i := 0; i < 5; i++ {
		randStr, _ := generateRandomString()
		ok, err := s.urlRepo.CreateURL(url, randStr)
		if err != nil {
			fmt.Println(err)
			return "Failed to create a short url", err
		}
		if !ok {
			continue
		}
		return hostURL + "/" + randStr, nil
	}
	return "", errors.New("failed to create a short url")
}

func generateRandomString() (string, error) {
	id, err := gonanoid.New(8)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	return id, nil
}
