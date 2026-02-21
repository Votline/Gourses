package db

import (
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB struct {
	log *zap.Logger
	db  *sqlx.DB
	bd  sq.StatementBuilderType
}

func NewDB(log *zap.Logger) (*DB, error) {
	var db *sqlx.DB
	var err error
	connStr := os.Getenv("POSTGRES_URL")
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err != nil {
			log.Error("Failed to connect to database", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(20)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(5 * time.Minute)

		log.Info("Connected to database")
		break
	}

	return &DB{
		log: log,
		db:  db,
		bd:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, err
}

func (d *DB) Close() error {
	return d.db.Close()
}
