package main

import (
	"flag"
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/config/server"
	serviceConfig "github.com/HimmelSpark/go-musthave-shortener.git/internal/config/shortener"
	storageConfig "github.com/HimmelSpark/go-musthave-shortener.git/internal/config/storage"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/handler"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/middleware"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
	"github.com/HimmelSpark/go-musthave-shortener.git/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)

	//sqlConfig := db.Init()
	//dbConn, err := db.GetDBConnection(sqlConfig)
	//if err != nil {
	//	panic(err)
	//}
	//urlRepo := repository.NewURLRepository(dbConn)

	serverConfig := server.Init()
	shortenerConfig := serviceConfig.Init()
	fileConfig, err := storageConfig.Init()
	if err != nil {
		panic(err)
	}

	flag.Parse()

	urlRepo, err := repository.NewFileURLRepository(fileConfig.FileStoragePath)
	if err != nil {
		panic(err)
	}

	shortenerService, err := service.NewShortenerService(urlRepo, shortenerConfig)
	if err != nil {
		panic(err)
	}
	shortenerHandler := handler.NewShortenerHandler(shortenerService)

	r.Post("/", shortenerHandler.ShortenURL)
	r.Post("/api/shorten", shortenerHandler.ShortenURLJSON)
	r.Get("/{urlId}", shortenerHandler.GetRedirectURL)

	if err := http.ListenAndServe(*serverConfig.ServerAddress, r); err != nil {
		panic(err)
	}
}
