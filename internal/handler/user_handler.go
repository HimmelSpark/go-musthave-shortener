package handler

import (
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/auth"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/model"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/mailru/easyjson"
)

type UserHandler struct {
	service  service.ShortenerService
	deletion service.DeletionService
}

func NewUserHandler(shortenerService service.ShortenerService, deletion service.DeletionService) *UserHandler {
	return &UserHandler{service: shortenerService, deletion: deletion}
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

func (h *UserHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	info := auth.FromContext(r.Context())
	if info.CookiePresent && !info.CookieValid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if info.UserID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var ids model.DeleteUserURLsRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &ids); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(ids) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	h.deletion.Submit(info.UserID, ids)
	w.WriteHeader(http.StatusAccepted)
}
