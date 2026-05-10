package repository

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

type urlRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
}

type fileURLRepository struct {
	mu            sync.RWMutex
	store         map[string]string
	byOriginalURL map[string]string
	byUserID      map[string][]string
	records       []urlRecord
	filePath      string
}

func NewFileURLRepository(filePath string) (URLRepository, error) {
	repo := &fileURLRepository{
		store:         make(map[string]string),
		byOriginalURL: make(map[string]string),
		byUserID:      make(map[string][]string),
		filePath:      filePath,
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (f *fileURLRepository) FindURLShortID(shortID string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.store[shortID], nil
}

func (f *fileURLRepository) CreateURL(originalURL string, shortID string, userID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if existingShortID, ok := f.byOriginalURL[originalURL]; ok {
		return false, &ErrDuplicateURL{ExistingShortID: existingShortID}
	}

	if _, exists := f.store[shortID]; exists {
		return false, nil
	}

	f.store[shortID] = originalURL
	f.byOriginalURL[originalURL] = shortID
	if userID != "" {
		f.byUserID[userID] = append(f.byUserID[userID], shortID)
	}

	rec := urlRecord{
		UUID:        strconv.Itoa(len(f.records) + 1),
		ShortURL:    shortID,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	f.records = append(f.records, rec)

	return true, f.save()
}

func (f *fileURLRepository) CreateURLBatch(items []URLBatchItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, item := range items {
		if _, exists := f.store[item.ShortID]; exists {
			return ErrBatchCollision
		}
	}

	for _, item := range items {
		f.store[item.ShortID] = item.OriginalURL
		f.byOriginalURL[item.OriginalURL] = item.ShortID
		if item.UserID != "" {
			f.byUserID[item.UserID] = append(f.byUserID[item.UserID], item.ShortID)
		}
		f.records = append(f.records, urlRecord{
			UUID:        strconv.Itoa(len(f.records) + 1),
			ShortURL:    item.ShortID,
			OriginalURL: item.OriginalURL,
			UserID:      item.UserID,
		})
	}

	return f.save()
}

func (f *fileURLRepository) GetUserURLs(userID string) ([]UserURL, error) {
	if userID == "" {
		return nil, nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()

	shortIDs := f.byUserID[userID]
	if len(shortIDs) == 0 {
		return nil, nil
	}
	result := make([]UserURL, 0, len(shortIDs))
	for _, sid := range shortIDs {
		if originalURL, ok := f.store[sid]; ok {
			result = append(result, UserURL{ShortID: sid, OriginalURL: originalURL})
		}
	}
	return result, nil
}

func (f *fileURLRepository) load() error {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}

	var records []urlRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	f.records = records
	for _, r := range records {
		f.store[r.ShortURL] = r.OriginalURL
		f.byOriginalURL[r.OriginalURL] = r.ShortURL
		if r.UserID != "" {
			f.byUserID[r.UserID] = append(f.byUserID[r.UserID], r.ShortURL)
		}
	}
	return nil
}

func (f *fileURLRepository) save() error {
	data, err := json.Marshal(f.records)
	if err != nil {
		return err
	}

	tmp := f.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, f.filePath)
}
