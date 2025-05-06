package db

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database предоставляет методы работы с базой данных
type Database struct {
	cluster *pgxpool.Pool
}

// NewDatabase создаёт новый экземпляр Database с переданным пулом соединений
func NewDataBase(cluster *pgxpool.Pool) *Database {
	return &Database{cluster: cluster}
}

// NewDb инициализирует пул соединений с БД по DSN и возвращает Database
func NewDb(ctx context.Context, dsn string) (*Database, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return NewDataBase(pool), nil
}

// Get выполняет один запрос и сканирует результат в dest
func (db *Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}
