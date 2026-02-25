package rdb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RDB struct {
	log *zap.Logger
	rdb *redis.Client
	ctx context.Context
}

func NewRDB(log *zap.Logger) (*RDB, error) {
	const op = "rdb.NewRDB"
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_SK_HOST") + ":" + os.Getenv("REDIS_SK_PORT"),
		Password: os.Getenv("REDIS_SK_PSWD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: ping redis: %w", op, err)
	}

	return &RDB{
		log: log,
		rdb: rdb,
		ctx: ctx,
	}, nil
}

func (r *RDB) Close() error {
	return r.rdb.Close()
}

func (r *RDB) NewSession(id, role string) (string, error) {
	const op = "UsersRedisDB.NewSession"

	sk := uuid.NewString()
	tx := r.rdb.TxPipeline()
	defer tx.Close()

	id = strings.TrimSpace(id)
	role = strings.TrimSpace(role)

	if err := tx.HSet(r.ctx, sk, map[string]string{
		"id":   id,
		"role": role,
	}).Err(); err != nil {
		return "", fmt.Errorf("%s: set hash: %w", op, err)
	}

	if err := tx.Expire(r.ctx, sk, 720*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("%s: expire hash: %w", op, err)
	}

	if _, err := tx.Exec(r.ctx); err != nil {
		return "", fmt.Errorf("%s: exec pipeline: %w", op, err)
	}

	return sk, nil
}

func (r *RDB) Validate(id, role, sk string) error {
	const op = "UsersRedisDB.Validate"

	id = strings.TrimSpace(id)
	role = strings.TrimSpace(role)

	fields, err := r.rdb.HGetAll(r.ctx, sk).Result()
	if err != nil {
		return fmt.Errorf("%q: get hash: %w", op, err)
	}
	if len(fields) == 0 {
		return fmt.Errorf("%q: hash is empty", op)
	}

	if fields["role"] != role {
		return fmt.Errorf("%s: invalid session", op)
	}

	if fields["id"] != id || fields["role"] != role {
		return fmt.Errorf("%q: invalid session", op)
	}

	return nil
}

func (r *RDB) Delete(sk string) error {
	const op = "UsersRedisDB.Delete"

	if err := r.rdb.Del(r.ctx, sk).Err(); err != nil {
		return fmt.Errorf("%s: delete hash: %w", op, err)
	}
	return nil
}
