package service

import (
	"errors"
	"strings"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/config/shortener"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type ErrConflict struct {
	ExistingURL string
}

func (e *ErrConflict) Error() string {
	return "url already shortened"
}

type ShortenBatchInput struct {
	CorrelationID string
	OriginalURL   string
}

type ShortenBatchOutput struct {
	CorrelationID string
	ShortURL      string
}

type UserURLOutput struct {
	ShortURL    string
	OriginalURL string
}

type ShortenerService interface {
	ShortenURL(url string, userID string) (string, error)
	FindURL(shortID string) (originalURL string, isDeleted bool, err error)
	ShortenURLBatch(items []ShortenBatchInput, userID string) ([]ShortenBatchOutput, error)
	GetUserURLs(userID string) ([]UserURLOutput, error)
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

func (s *shortenerService) FindURL(shortID string) (string, bool, error) {
	shortID = strings.TrimSpace(shortID)
	return s.urlRepo.FindURLShortID(shortID)
}

func (s *shortenerService) ShortenURL(url string, userID string) (string, error) {
	url = strings.TrimSpace(url)
	baseURL := strings.TrimRight(*s.serviceConfig.BaseURL, "/")
	for i := 0; i < 5; i++ {
		randStr, _ := generateRandomString()
		ok, err := s.urlRepo.CreateURL(url, randStr, userID)
		if err != nil {
			var dupErr *repository.ErrDuplicateURL
			if errors.As(err, &dupErr) {
				return "", &ErrConflict{
					ExistingURL: baseURL + "/" + dupErr.ExistingShortID,
				}
			}
			return "Failed to create a short url", err
		}
		if !ok {
			continue
		}
		return baseURL + "/" + randStr, nil
	}
	return "", errors.New("failed to create a short url")
}

func (s *shortenerService) ShortenURLBatch(items []ShortenBatchInput, userID string) ([]ShortenBatchOutput, error) {
	baseURL := strings.TrimRight(*s.serviceConfig.BaseURL, "/")

	for attempt := 0; attempt < 5; attempt++ {
		batchItems := make([]repository.URLBatchItem, len(items))
		for i, item := range items {
			randStr, _ := generateRandomString()
			batchItems[i] = repository.URLBatchItem{
				OriginalURL: strings.TrimSpace(item.OriginalURL),
				ShortID:     randStr,
				UserID:      userID,
			}
		}

		err := s.urlRepo.CreateURLBatch(batchItems)
		if err != nil {
			if errors.Is(err, repository.ErrBatchCollision) {
				continue
			}
			return nil, err
		}

		result := make([]ShortenBatchOutput, len(items))
		for i, item := range items {
			result[i] = ShortenBatchOutput{
				CorrelationID: item.CorrelationID,
				ShortURL:      baseURL + "/" + batchItems[i].ShortID,
			}
		}
		return result, nil
	}

	return nil, errors.New("failed to create batch short urls: too many collisions")
}

func (s *shortenerService) GetUserURLs(userID string) ([]UserURLOutput, error) {
	items, err := s.urlRepo.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	baseURL := strings.TrimRight(*s.serviceConfig.BaseURL, "/")
	result := make([]UserURLOutput, len(items))
	for i, it := range items {
		result[i] = UserURLOutput{
			ShortURL:    baseURL + "/" + it.ShortID,
			OriginalURL: it.OriginalURL,
		}
	}
	return result, nil
}

func generateRandomString() (string, error) {
	id, err := gonanoid.New(8)
	if err != nil {
		panic(err)
	}
	return id, nil
}
