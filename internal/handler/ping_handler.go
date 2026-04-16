package handler

import (
	"database/sql"
	"net/http"
)

type PingHandler struct {
	db *sql.DB
}

func NewPingHandler(db *sql.DB) *PingHandler {
	return &PingHandler{db: db}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}

	if err := h.db.PingContext(r.Context()); err != nil {
		http.Error(w, "database ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
