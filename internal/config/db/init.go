package db

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	time2 "time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLConfig struct {
	Driver                 *string
	Dsn                    *string
	MaxOpenConnections     *int
	MaxIdleConnections     *int
	ConnectionAwaitTimeout *time2.Duration
}

func Init() *SQLConfig {
	cfg := &SQLConfig{
		Driver:                 flag.String("db-driver", "pgx", "Database driver"),
		Dsn:                    flag.String("d", "", "Database DSN"),
		MaxOpenConnections:     flag.Int("db-max-open", 10, "Maximum number of open connections to the database"),
		MaxIdleConnections:     flag.Int("db-max-idle", 5, "Maximum number of idle connections to the database"),
		ConnectionAwaitTimeout: flag.Duration("db-time", 90*time2.Second, "Maximum number of seconds to wait for connections to the database"),
	}

	return cfg
}

func (c *SQLConfig) ApplyEnv() {
	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		*c.Dsn = dsn
	}
}

func GetDBConnection(config *SQLConfig) (*sql.DB, error) {
	db, err := sql.Open(*config.Driver, *config.Dsn)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return nil, err
	}
	db.SetMaxOpenConns(*config.MaxOpenConnections)
	db.SetMaxIdleConns(*config.MaxIdleConnections)
	db.SetConnMaxLifetime(*config.ConnectionAwaitTimeout)

	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil
}
