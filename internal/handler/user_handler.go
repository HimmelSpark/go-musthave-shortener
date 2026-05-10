package handler

import (
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/auth"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/model"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/mailru/easyjson"
)

type UserHandler struct {
	service service.ShortenerService
}

func NewUserHandler(shortenerService service.ShortenerService) *UserHandler {
	return &UserHandler{shortenerService}
}

func (h *UserHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	info := auth.FromContext(r.Context())
	if info.CookiePresent && !info.CookieValid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	items, err := h.service.GetUserURLs(info.UserID)
	if err != nil {
		http.Error(w, "failed to get user urls", http.StatusInternalServerError)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make(model.UserURLList, len(items))
	for i, it := range items {
		resp[i] = model.UserURLItem{
			ShortURL:    it.ShortURL,
			OriginalURL: it.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = easyjson.MarshalToWriter(&resp, w)
}
