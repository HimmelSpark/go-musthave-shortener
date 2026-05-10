package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	serverCfg := server.Init()
	shortenerCfg := serviceConfig.Init()
	sqlCfg := dbConfig.Init()
	authCfg := authConfig.Init()
	fileCfg, err := storageConfig.Init()
	if err != nil {
		panic(err)
	}

	flag.Parse()
	sqlCfg.ApplyEnv()

	dbConn, urlRepo := initStorage(sqlCfg, fileCfg)

	shortenerSvc, err := service.NewShortenerService(urlRepo, shortenerCfg)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deletionSvc := service.NewDeletionService(urlRepo, 1024, 256, time.Second)
	go deletionSvc.Run(ctx)

	router := buildRouter(*authCfg.SecretKey, shortenerSvc, deletionSvc, dbConn)
	runHTTPServer(ctx, *serverCfg.ServerAddress, router)
}

func buildRouter(authSecret string, shortenerSvc service.ShortenerService, deletionSvc service.DeletionService, dbConn *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)
	r.Use(auth.Middleware(authSecret))

	shortenerH := handler.NewShortenerHandler(shortenerSvc)
	pingH := handler.NewPingHandler(dbConn)
	userH := handler.NewUserHandler(shortenerSvc, deletionSvc)

	r.Post("/", shortenerH.ShortenURL)
	r.Post("/api/shorten", shortenerH.ShortenURLJSON)
	r.Post("/api/shorten/batch", shortenerH.ShortenURLBatch)
	r.Get("/api/user/urls", userH.GetUserURLs)
	r.Delete("/api/user/urls", userH.DeleteUserURLs)
	r.Get("/ping", pingH.Ping)
	r.Get("/{urlId}", shortenerH.GetRedirectURL)

	return r
}

func runHTTPServer(ctx context.Context, addr string, h http.Handler) {
	srv := &http.Server{Addr: addr, Handler: h}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
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
