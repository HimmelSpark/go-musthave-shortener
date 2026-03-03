package main

import (
	"flag"
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/config/server"
	serviceConfig "github.com/HimmelSpark/go-musthave-shortener.git/internal/config/shortener"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/handler"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/middleware"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)

	//sqlConfig := db.Init()
	//dbConn := db.GetDBConnection(sqlConfig)
	//urlRepo := repository.NewURLRepository(dbConn)

	serverConfig := server.Init()
	shortenerConfig := serviceConfig.Init()

	flag.Parse()

	urlRepo := repository.NewInMemoryURLRepository()

	shortenerService := service.NewShortenerService(urlRepo, shortenerConfig)
	shortenerHandler := handler.NewShortenerHandler(shortenerService)

	r.Post("/", shortenerHandler.ShortenURL)
	r.Get("/{urlId}", shortenerHandler.GetRedirectURL)

	if err := http.ListenAndServe(*serverConfig.ServerAddress, r); err != nil {
		panic(err)
	}
}
