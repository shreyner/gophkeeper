package database

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/net/context"
)

func NewDataBase(ctx context.Context, dburi string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dburi)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
