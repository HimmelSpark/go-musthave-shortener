package service

import (
	"errors"
	"strings"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/config/shortener"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type ShortenerService interface {
	ShortenURL(url string) (string, error)
	FindURL(shortID string) (string, error)
}
type shortenerService struct {
	urlRepo       repository.URLRepository
	serviceConfig *shortener.ServiceConfig
}

func NewShortenerService(urlRepo repository.URLRepository, serviceConfig *shortener.ServiceConfig) (ShortenerService, error) {
	if serviceConfig == nil || serviceConfig.BaseURL == nil {
		return nil, errors.New("shortener service config is nil")
	}
	if strings.TrimSpace(*serviceConfig.BaseURL) == "" {
		return nil, errors.New("shortener service config is empty")
	}

	return &shortenerService{
		urlRepo:       urlRepo,
		serviceConfig: serviceConfig,
	}, nil
}

func (s *shortenerService) FindURL(shortID string) (string, error) {
	shortID = strings.TrimSpace(shortID)
	return s.urlRepo.FindURLShortID(shortID)
}

func (s *shortenerService) ShortenURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	baseURL := strings.TrimRight(*s.serviceConfig.BaseURL, "/")
	for i := 0; i < 5; i++ {
		randStr, _ := generateRandomString()
		ok, err := s.urlRepo.CreateURL(url, randStr)
		if err != nil {
			return "Failed to create a short url", err
		}
		if !ok {
			continue
		}
		return baseURL + "/" + randStr, nil
	}
	return "", errors.New("failed to create a short url")
}

func generateRandomString() (string, error) {
	id, err := gonanoid.New(8)
	if err != nil {
		panic(err)
	}
	return id, nil
}
