package repository

type inMemoryURLRepository struct {
	store         map[string]string
	byOriginalURL map[string]string
}

func NewInMemoryURLRepository() URLRepository {
	return &inMemoryURLRepository{
		store:         make(map[string]string),
		byOriginalURL: make(map[string]string),
	}
}

func (i *inMemoryURLRepository) FindURLShortID(s string) (string, error) {
	redirectURL, ok := i.store[s]
	if !ok {
		return "", nil
	}
	return redirectURL, nil
}

func (i *inMemoryURLRepository) CreateURL(originalURL string, shortID string) (bool, error) {
	if existingShortID, ok := i.byOriginalURL[originalURL]; ok {
		return false, &ErrDuplicateURL{ExistingShortID: existingShortID}
	}

	i.store[shortID] = originalURL
	i.byOriginalURL[originalURL] = shortID
	return true, nil
}

func (i *inMemoryURLRepository) CreateURLBatch(items []URLBatchItem) error {
	for _, item := range items {
		i.store[item.ShortID] = item.OriginalURL
		i.byOriginalURL[item.OriginalURL] = item.ShortID
	}
	return nil
}
