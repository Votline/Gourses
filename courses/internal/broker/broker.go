package broker

import (
	"context"
	"fmt"
	"os"

	gc "courses/internal/gracefulshutdown"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Broker struct {
	log     *zap.Logger
	channel *redis.Client
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
	}, nil
}

func (b *Broker) Close(ctx context.Context) error {
	return gc.Shutdown(b.channel.Close, ctx)
}

func (b *Broker) Subscribe(ctx context.Context, channel string) <-chan string {
	const op = "broker.Subscribe"

	pubsub := b.channel.Subscribe(ctx, channel)

	msgChan := make(chan string)

	go func() {
		defer pubsub.Close()
		defer close(msgChan)

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				b.log.Error(op, zap.Error(err))
				return
			}

			msgChan <- msg.Payload
		}
	}()

	return msgChan
}
