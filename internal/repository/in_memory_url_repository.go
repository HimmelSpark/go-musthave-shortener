package repository

type inMemoryURLRepository struct {
	store         map[string]string
	byOriginalURL map[string]string
	byUserID      map[string][]string
}

func NewInMemoryURLRepository() URLRepository {
	return &inMemoryURLRepository{
		store:         make(map[string]string),
		byOriginalURL: make(map[string]string),
		byUserID:      make(map[string][]string),
	}
}

func (i *inMemoryURLRepository) FindURLShortID(s string) (string, error) {
	redirectURL, ok := i.store[s]
	if !ok {
		return "", nil
	}
	return redirectURL, nil
}

func (i *inMemoryURLRepository) CreateURL(originalURL string, shortID string, userID string) (bool, error) {
	if existingShortID, ok := i.byOriginalURL[originalURL]; ok {
		return false, &ErrDuplicateURL{ExistingShortID: existingShortID}
	}

	i.store[shortID] = originalURL
	i.byOriginalURL[originalURL] = shortID
	if userID != "" {
		i.byUserID[userID] = append(i.byUserID[userID], shortID)
	}
	return true, nil
}

func (i *inMemoryURLRepository) CreateURLBatch(items []URLBatchItem) error {
	for _, item := range items {
		i.store[item.ShortID] = item.OriginalURL
		i.byOriginalURL[item.OriginalURL] = item.ShortID
		if item.UserID != "" {
			i.byUserID[item.UserID] = append(i.byUserID[item.UserID], item.ShortID)
		}
	}
	return nil
}

func (i *inMemoryURLRepository) GetUserURLs(userID string) ([]UserURL, error) {
	if userID == "" {
		return nil, nil
	}
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
