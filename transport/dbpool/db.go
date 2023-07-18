package dbpool

import (
	"database/sql"
	
	_ "github.com/lib/pq"
)

func NewDBPostgresPool(dsn string) (*sql.DB, error) {
	pool, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	err = pool.Ping()
	if err != nil {
		return nil, err
	}
	return pool, nil
}
