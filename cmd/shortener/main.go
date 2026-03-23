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
	//dbConn, err := db.GetDBConnection(sqlConfig)
	//if err != nil {
	//	panic(err)
	//}
	//urlRepo := repository.NewURLRepository(dbConn)

	serverConfig := server.Init()
	shortenerConfig := serviceConfig.Init()

	flag.Parse()

	urlRepo := repository.NewInMemoryURLRepository()

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
