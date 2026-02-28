package main

import (
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/config/db"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/handler"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/middleware"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
)

func main() {
	mux := http.NewServeMux()

	dbConn := db.GetDbConnection()
	urlRepo := repository.NewUrlRepository(dbConn)

	shortenerService := service.NewShortenerService(urlRepo)
	shortenerHandler := handler.NewShortenerHandler(shortenerService)

	mux.HandleFunc("POST /", shortenerHandler.ShortenUrl)
	mux.HandleFunc("GET /{urlId}", shortenerHandler.GetRedirectURL)

	handlerChain := middleware.LoggingMiddleware(mux)

	if err := http.ListenAndServe("localhost:8080", handlerChain); err != nil {
		panic(err)
	}
}
