package database

import (
	"context"
	"fmt"
	"os"
	"time"
	"trade/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Queries *db.Queries
	DBPool  *pgxpool.Pool
}

func New() (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbURL := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("error creating the connection pool: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %v", err)
	}

	queries := db.New(pool)

	db := Database{
		Queries: queries,
		DBPool:  pool,
	}

	return &db, nil
}
