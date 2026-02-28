package main

import (
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/handler"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/middleware"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
)

func main() {
	mux := http.NewServeMux()

	//dbConn := db.GetDBConnection()
	//urlRepo := repository.NewURLRepository(dbConn)
	urlRepo := repository.NewInMemoryURLRepository()

	shortenerService := service.NewShortenerService(urlRepo)
	shortenerHandler := handler.NewShortenerHandler(shortenerService)

	mux.HandleFunc("POST /", shortenerHandler.ShortenURL)
	mux.HandleFunc("GET /{urlId}", shortenerHandler.GetRedirectURL)

	handlerChain := middleware.LoggingMiddleware(mux)

	if err := http.ListenAndServe("localhost:8080", handlerChain); err != nil {
		panic(err)
	}
}
