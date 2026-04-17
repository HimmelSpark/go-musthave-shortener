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
}

type fileURLRepository struct {
	mu       sync.RWMutex
	store    map[string]string
	records  []urlRecord
	filePath string
}

func NewFileURLRepository(filePath string) (URLRepository, error) {
	repo := &fileURLRepository{
		store:    make(map[string]string),
		filePath: filePath,
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

func (f *fileURLRepository) CreateURL(originalURL string, shortID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for existingShortID, existingOrigURL := range f.store {
		if existingOrigURL == originalURL {
			return false, &ErrDuplicateURL{ExistingShortID: existingShortID}
		}
	}

	if _, exists := f.store[shortID]; exists {
		return false, nil
	}

	f.store[shortID] = originalURL

	rec := urlRecord{
		UUID:        strconv.Itoa(len(f.records) + 1),
		ShortURL:    shortID,
		OriginalURL: originalURL,
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
		f.records = append(f.records, urlRecord{
			UUID:        strconv.Itoa(len(f.records) + 1),
			ShortURL:    item.ShortID,
			OriginalURL: item.OriginalURL,
		})
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
