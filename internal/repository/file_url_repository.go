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
	IsDeleted   bool   `json:"is_deleted,omitempty"`
}

type fileURLRepository struct {
	mu            sync.RWMutex
	store         map[string]string
	byOriginalURL map[string]string
	byUserID      map[string][]string
	byShortIDUser map[string]string
	deleted       map[string]bool
	records       []urlRecord
	filePath      string
}

func NewFileURLRepository(filePath string) (URLRepository, error) {
	repo := &fileURLRepository{
		store:         make(map[string]string),
		byOriginalURL: make(map[string]string),
		byUserID:      make(map[string][]string),
		byShortIDUser: make(map[string]string),
		deleted:       make(map[string]bool),
		filePath:      filePath,
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (f *fileURLRepository) FindURLShortID(shortID string) (string, bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	originalURL, ok := f.store[shortID]
	if !ok {
		return "", false, nil
	}
	return originalURL, f.deleted[shortID], nil
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
		f.byShortIDUser[shortID] = userID
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
			f.byShortIDUser[item.ShortID] = item.UserID
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

func (f *fileURLRepository) DeleteUserURLs(userID string, shortIDs []string) error {
	if userID == "" || len(shortIDs) == 0 {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	changed := false
	for _, sid := range shortIDs {
		if f.byShortIDUser[sid] != userID || f.deleted[sid] {
			continue
		}
		f.deleted[sid] = true
		for i := range f.records {
			if f.records[i].ShortURL == sid {
				f.records[i].IsDeleted = true
				break
			}
		}
		changed = true
	}
	if !changed {
		return nil
	}
	return f.save()
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
			f.byShortIDUser[r.ShortURL] = r.UserID
		}
		if r.IsDeleted {
			f.deleted[r.ShortURL] = true
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
