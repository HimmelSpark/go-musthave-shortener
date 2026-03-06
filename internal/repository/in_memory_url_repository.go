package repository

type inMemoryURLRepository struct {
	store map[string]string
}

func NewInMemoryURLRepository() URLRepository {
	return &inMemoryURLRepository{store: make(map[string]string)}
}

func (i inMemoryURLRepository) FindURLShortID(s string) (string, error) {
	redirectURL, ok := i.store[s]
	if !ok {
		return "", nil
	}
	return redirectURL, nil
}

func (i inMemoryURLRepository) CreateURL(originalURL string, shortID string) (bool, error) {
	i.store[shortID] = originalURL
	return true, nil
}
