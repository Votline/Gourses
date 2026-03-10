package db

import (
	"fmt"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	db *sqlx.DB
	bd sq.StatementBuilderType
}

type CourseInfo struct {
	Name  string `db:"name"`
	Desc  string `db:"description"`
	Price string `db:"price"`
}

func NewDB(log *zap.Logger) (*DB, error) {
	var db *sqlx.DB
	var err error

	connStr := os.Getenv("POSTGRES_URL")
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err != nil {
			log.Error("failed to connect to db", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		break
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return &DB{
		db: db,
		bd: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) NewCourse(id, name, desc, price, userID string) error {
	const op = "db.NewCourse"

	query, args, err := d.bd.Insert("courses").
		Columns("id", "name", "description", "price", "user_id").
		Values(id, name, desc, price, userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: create query: %w", op, err)
	}

	if _, err := d.db.Exec(query, args...); err != nil {
		return fmt.Errorf("%s: new course: %w", op, err)
	}

	return nil
}

func (d *DB) GetCourse(id string) (*CourseInfo, error) {
	const op = "db.GetCourses"

	query, args, err := d.bd.Select("name", "description", "price").
		From("courses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: get query: %w", op, err)
	}

	courseInfo := CourseInfo{}
	if err := d.db.Get(&courseInfo, query, args...); err != nil {
		return nil, fmt.Errorf("%s: get course: %w", op, err)
	}

	return &courseInfo, nil
}

func (d *DB) UpdateCourse(userID, userRole, id, name, desc, price string) error {
	const op = "db.UpdateCourse"

	q := d.bd.Update("courses").
		Set("name", name).
		Set("description", desc).
		Set("price", price).
		Where(sq.Eq{"id": id})

	if userRole != "admin" {
		q = q.Where(sq.Eq{"user_id": userID})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("%s: update query: %w", op, err)
	}

	res, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: update course: %w", op, err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%s: update course: %w", op, err)
	} else if n == 0 {
		return fmt.Errorf("%s: no course found", op)
	}

	return nil
}

func (d *DB) DeleteCourse(id, userID, userRole string) error {
	const op = "db.DeleteCourse"

	q := d.bd.Delete("courses").Where(sq.Eq{"id": id})

	if userRole != "admin" {
		q = q.Where(sq.Eq{"user_id": userID})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("%s: delete query: %w", op, err)
	}

	res, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: delete course: %w", op, err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%s: delete course: %w", op, err)
	} else if n == 0 {
		return fmt.Errorf("%s: no course found", op)
	}

	return nil
}

func (d *DB) DeleteCourseByID(userID string) error {
	const op = "db.DeleteCouseByID"

	query, args, err := d.bd.Delete("courses").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: delete query: %w", op, err)
	}

	res, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: delete course: %w", op, err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%s: delete course: %w", op, err)
	} else if n == 0 {
		return fmt.Errorf("%s: no course found", op)
	}

	return nil
}
