package broker

import (
	"context"
	"fmt"
	"os"

	gc "users/internal/gracefulshutdown"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Broker struct {
	log     *zap.Logger
	channel *redis.Client
	ctx     context.Context
}

func NewBroker(log *zap.Logger) (*Broker, error) {
	const op = "broker.NewBroker"

	channel := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_BK_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_BK_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := channel.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: ping redis: %w", op, err)
	}

	return &Broker{
		log:     log,
		channel: channel,
		ctx:     ctx,
	}, nil
}

func (b *Broker) Close(ctx context.Context) error {
	return gc.Shutdown(b.channel.Close, ctx)
}

func (b *Broker) Publish(channel, message string) error {
	const op = "broker.Publish"

	if err := b.channel.Publish(b.ctx, channel, message).Err(); err != nil {
		return fmt.Errorf("%s: publish message: %w", op, err)
	}
	return nil
}
