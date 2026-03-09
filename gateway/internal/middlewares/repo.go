package middlewares

import (
	"context"
	"fmt"
	"os"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

type UserInfo struct {
	ID    string `json:"id"`
	Role  string `json:"role"`
	token *jwt.Token
	jwt.RegisteredClaims
}

type Mdwr struct {
	validate func(ctx context.Context, tokenStr, sessionKey string) (*pb.ValidateRes, error)
	rdb      *redis.Client
	ctx      context.Context
}

func NewMdwr(validate func(ctx context.Context, tokenStr, sessionKey string) (*pb.ValidateRes, error)) (*Mdwr, error) {
	const op = "middlewares.NewMdwr"

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_RL_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_RL_PASSWORD"),
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: redis ping: %w", op, err)
	}
	return &Mdwr{
		validate: validate,
		ctx:      ctx,
		rdb:      rdb,
	}, nil
}
