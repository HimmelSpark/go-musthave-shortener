package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
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

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host
	baseURL := scheme + "://" + host

	short, err := h.service.ShortenURL(url, baseURL)

	// todo продумать бизнесовые ошибки и ловить их тут
	if err != nil {
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
	}

	if url == "" {
		http.NotFound(w, r)
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
