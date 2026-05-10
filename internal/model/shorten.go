package model

//easyjson:json
type ShortenRequest struct {
	URL string `json:"url"`
}

//easyjson:json
type ShortenResponse struct {
	Result string `json:"result"`
}

//easyjson:json
type BatchShortenRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type BatchShortenRequest []BatchShortenRequestItem

//easyjson:json
type BatchShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchShortenResponse []BatchShortenResponseItem

//easyjson:json
type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//easyjson:json
type UserURLList []UserURLItem
