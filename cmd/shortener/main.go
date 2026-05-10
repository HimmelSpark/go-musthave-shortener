package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/auth"
	authConfig "github.com/HimmelSpark/go-musthave-shortener.git/internal/config/auth"
	dbConfig "github.com/HimmelSpark/go-musthave-shortener.git/internal/config/db"
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
	serverConfig := server.Init()
	shortenerConfig := serviceConfig.Init()
	sqlConfig := dbConfig.Init()
	authCfg := authConfig.Init()
	fileConfig, err := storageConfig.Init()
	if err != nil {
		panic(err)
	}

	flag.Parse()

	sqlConfig.ApplyEnv()

	dbConn, urlRepo := initStorage(sqlConfig, fileConfig)

	shortenerService, err := service.NewShortenerService(urlRepo, shortenerConfig)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)
	r.Use(auth.Middleware(*authCfg.SecretKey))

	shortenerHandler := handler.NewShortenerHandler(shortenerService)
	pingHandler := handler.NewPingHandler(dbConn)
	userHandler := handler.NewUserHandler(shortenerService)

	r.Post("/", shortenerHandler.ShortenURL)
	r.Post("/api/shorten", shortenerHandler.ShortenURLJSON)
	r.Post("/api/shorten/batch", shortenerHandler.ShortenURLBatch)
	r.Get("/api/user/urls", userHandler.GetUserURLs)
	r.Get("/ping", pingHandler.Ping)
	r.Get("/{urlId}", shortenerHandler.GetRedirectURL)

	if err := http.ListenAndServe(*serverConfig.ServerAddress, r); err != nil {
		panic(err)
	}
}

func initStorage(sqlConfig *dbConfig.SQLConfig, fileConfig *storageConfig.Config) (*sql.DB, repository.URLRepository) {
	if *sqlConfig.Dsn != "" {
		dbConn, err := dbConfig.GetDBConnection(sqlConfig)
		if err != nil {
			log.Printf("Failed to connect to database: %v. Falling back to file/in-memory storage.", err)
		} else {
			if err := dbConfig.RunMigrations(*sqlConfig.Dsn); err != nil {
				log.Printf("Failed to run migrations: %v", err)
			}
			log.Println("Using PostgreSQL storage")
			return dbConn, repository.NewURLRepository(dbConn)
		}
	}

	if fileConfig.FileStoragePath != "" {
		repo, err := repository.NewFileURLRepository(fileConfig.FileStoragePath)
		if err != nil {
			log.Printf("Failed to init file storage: %v. Falling back to in-memory storage.", err)
		} else {
			log.Println("Using file storage")
			return nil, repo
		}
	}

	log.Println("Using in-memory storage")
	return nil, repository.NewInMemoryURLRepository()
}
