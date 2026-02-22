package db

import (
	"fmt"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB struct {
	db *sqlx.DB
	bd sq.StatementBuilderType
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
		db: db,
		bd: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, err
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) RegUser(id, username, role, pswd string) error {
	const op = "db.RegUser"

	query, args, err := d.bd.Insert("users").
		Columns("id, user_name", "role", "password").
		Values(id, username, role, pswd).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: create query: %w", op, err)
	}

	if _, err := d.db.Exec(query, args...); err != nil {
		return fmt.Errorf("%s: insert user: %w", op, err)
	}

	return nil
}
