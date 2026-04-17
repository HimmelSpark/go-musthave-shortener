package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/model"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/mailru/easyjson"
)

type ShortenerHandler struct {
	service service.ShortenerService
}

func NewShortenerHandler(shortenerService service.ShortenerService) *ShortenerHandler {
	return &ShortenerHandler{shortenerService}
}

func (h *ShortenerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	url := strings.TrimSpace(string(body))
	if url == "" {
		http.Error(w, "Empty URL", http.StatusBadRequest)
		return
	}

	short, err := h.service.ShortenURL(url)
	if err != nil {
		var conflictErr *service.ErrConflict
		if errors.As(err, &conflictErr) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(conflictErr.ExistingURL))
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(short))
}

func (h *ShortenerHandler) GetRedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlShortID := r.PathValue("urlId")
	if urlShortID == "" {
		http.NotFound(w, r)
		return
	}

	url, err := h.service.FindURL(urlShortID)
	if err != nil {
		http.Error(w, "Failed to find url", http.StatusBadRequest)
		return
	}

	if url == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	url := strings.TrimSpace(req.URL)
	if url == "" {
		http.Error(w, "empty URL", http.StatusBadRequest)
		return
	}

	short, err := h.service.ShortenURL(url)
	if err != nil {
		var conflictErr *service.ErrConflict
		if errors.As(err, &conflictErr) {
			resp := model.ShortenResponse{Result: conflictErr.ExistingURL}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = easyjson.MarshalToWriter(&resp, w)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := model.ShortenResponse{Result: short}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = easyjson.MarshalToWriter(&resp, w)
}

func (h *ShortenerHandler) ShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	var req model.BatchShortenRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req) == 0 {
		http.Error(w, "empty batch", http.StatusBadRequest)
		return
	}

	inputs := make([]service.ShortenBatchInput, len(req))
	for i, item := range req {
		url := strings.TrimSpace(item.OriginalURL)
		if url == "" {
			http.Error(w, "empty URL in batch", http.StatusBadRequest)
			return
		}
		inputs[i] = service.ShortenBatchInput{
			CorrelationID: item.CorrelationID,
			OriginalURL:   url,
		}
	}

	outputs, err := h.service.ShortenURLBatch(inputs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make(model.BatchShortenResponse, len(outputs))
	for i, o := range outputs {
		resp[i] = model.BatchShortenResponseItem{
			CorrelationID: o.CorrelationID,
			ShortURL:      o.ShortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = easyjson.MarshalToWriter(&resp, w)
}
