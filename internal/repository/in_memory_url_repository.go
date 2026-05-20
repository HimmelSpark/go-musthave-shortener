package repository

import "sync"

type inMemoryURLRepository struct {
	mu            sync.RWMutex
	store         map[string]string
	byOriginalURL map[string]string
	byUserID      map[string][]string
	byShortIDUser map[string]string
	deleted       map[string]bool
}

func NewInMemoryURLRepository() URLRepository {
	return &inMemoryURLRepository{
		store:         make(map[string]string),
		byOriginalURL: make(map[string]string),
		byUserID:      make(map[string][]string),
		byShortIDUser: make(map[string]string),
		deleted:       make(map[string]bool),
	}
}

func (i *inMemoryURLRepository) FindURLShortID(s string) (string, bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	redirectURL, ok := i.store[s]
	if !ok {
		return "", false, nil
	}
	return redirectURL, i.deleted[s], nil
}

func (i *inMemoryURLRepository) CreateURL(originalURL string, shortID string, userID string) (bool, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if existingShortID, ok := i.byOriginalURL[originalURL]; ok {
		return false, &ErrDuplicateURL{ExistingShortID: existingShortID}
	}

	i.store[shortID] = originalURL
	i.byOriginalURL[originalURL] = shortID
	if userID != "" {
		i.byUserID[userID] = append(i.byUserID[userID], shortID)
		i.byShortIDUser[shortID] = userID
	}
	return true, nil
}

func (i *inMemoryURLRepository) CreateURLBatch(items []URLBatchItem) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, item := range items {
		i.store[item.ShortID] = item.OriginalURL
		i.byOriginalURL[item.OriginalURL] = item.ShortID
		if item.UserID != "" {
			i.byUserID[item.UserID] = append(i.byUserID[item.UserID], item.ShortID)
			i.byShortIDUser[item.ShortID] = item.UserID
		}
	}
	return nil
}

func (i *inMemoryURLRepository) GetUserURLs(userID string) ([]UserURL, error) {
	if userID == "" {
		return nil, nil
	}
	i.mu.RLock()
	defer i.mu.RUnlock()

	shortIDs := i.byUserID[userID]
	if len(shortIDs) == 0 {
		return nil, nil
	}
	result := make([]UserURL, 0, len(shortIDs))
	for _, sid := range shortIDs {
		if originalURL, ok := i.store[sid]; ok {
			result = append(result, UserURL{ShortID: sid, OriginalURL: originalURL})
		}
	}
	return result, nil
}

func (i *inMemoryURLRepository) DeleteUserURLs(userID string, shortIDs []string) error {
	if userID == "" || len(shortIDs) == 0 {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, sid := range shortIDs {
		if i.byShortIDUser[sid] == userID {
			i.deleted[sid] = true
		}
	}
	return nil
}
