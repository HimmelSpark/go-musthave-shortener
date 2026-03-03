package main

import (
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/handler"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/middleware"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)

	//dbConn := db.GetDBConnection()
	//urlRepo := repository.NewURLRepository(dbConn)
	urlRepo := repository.NewInMemoryURLRepository()

	shortenerService := service.NewShortenerService(urlRepo)
	shortenerHandler := handler.NewShortenerHandler(shortenerService)

	r.Post("/", shortenerHandler.ShortenURL)
	r.Get("/{urlId}", shortenerHandler.GetRedirectURL)

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}
}
