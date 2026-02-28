package db

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	time2 "time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func GetDbConnection() *sql.DB {
	driver := flag.String("db-driver", "pgx", "Database driver")
	dsn := flag.String("dsn", "", "Database DSN")
	maxOpen := flag.Int("db-max-open", 10, "Maximum number of open connections to the database")
	maxIdle := flag.Int("db-max-idle", 5, "Maximum number of idle connections to the database")
	time := flag.Duration("db-time", 90*time2.Second, "Maximum number of seconds to wait for connections to the database")

	flag.Parse()

	if *dsn == "" {
		fmt.Println("dsn is required")
	}

	db, err := sql.Open(*driver, *dsn)
	if err != nil {
		fmt.Println("Error opening database:", err)
	}

	db.SetMaxOpenConns(*maxOpen)
	db.SetMaxIdleConns(*maxIdle)
	db.SetConnMaxLifetime(*time)

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}

	return db
}
