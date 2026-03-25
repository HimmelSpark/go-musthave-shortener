package repository

import (
	"bufio"
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
	counter  int
	file     *os.File
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

	if _, exists := f.store[shortID]; exists {
		return false, nil
	}

	f.store[shortID] = originalURL
	f.counter++

	rec := urlRecord{
		UUID:        strconv.Itoa(f.counter),
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	return true, f.append(rec)
}

func (f *fileURLRepository) load() error {
	file, err := os.OpenFile(f.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var rec urlRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			file.Close()
			return err
		}
		f.store[rec.ShortURL] = rec.OriginalURL
		f.counter++
	}
	if err := scanner.Err(); err != nil {
		file.Close()
		return err
	}

	f.file = file
	return nil
}

func (f *fileURLRepository) append(rec urlRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = f.file.Write(data)
	return err
}
