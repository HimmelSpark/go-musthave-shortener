package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type ShortenerService interface {
	ShortenURL(url string, hostUrl string) (string, error)
	FindUrl(shortId string) (string, error)
}
type shortenerService struct {
	urlRepo repository.UrlRepository
}

func NewShortenerService(urlRepo repository.UrlRepository) ShortenerService {
	return &shortenerService{urlRepo: urlRepo}
}

func (s *shortenerService) FindUrl(shortId string) (string, error) {
	shortId = strings.TrimSpace(shortId)
	url, err := s.urlRepo.FindUrl(shortId)
	return url, err
}

func (s *shortenerService) ShortenURL(url string, hostUrl string) (string, error) {
	url = strings.TrimSpace(url)
	for i := 0; i < 5; i++ {
		randStr, _ := generateRandomString()
		ok, err := s.urlRepo.CreateUrl(url, randStr)
		if err != nil {
			fmt.Println(err)
			return "Failed to create a short url", err
		}
		if !ok {
			continue
		}
		return hostUrl + "/" + randStr, nil
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
