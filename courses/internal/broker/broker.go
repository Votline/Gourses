package broker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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

func (b *Broker) ListenStream(ctx context.Context, stream, group, consumer string) (<-chan string, error) {
	const op = "broker.Subscribe"

	msgChan := make(chan string)

	err := b.channel.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && strings.Contains(err.Error(), "BUSYGROUP") {
		err := b.channel.XGroupDelConsumer(ctx, stream, group, consumer).Err()
		if err != nil {
			return nil, fmt.Errorf("%s: delete consumer: %w", op, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("%s: create stream: %w", op, err)
	}

	go func() {
		defer close(msgChan)

		currentID := "0"
		currentBlock := 100

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			entries, err := b.channel.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, currentID},
				Count:    1,
				Block:    time.Duration(currentBlock) * time.Millisecond,
			}).Result()

			if (err == redis.Nil || (len(entries) > 0 && len(entries[0].Messages) == 0)) && currentID == "0" {
				currentID = ">"
				currentBlock = 0
				continue
			}

			if err != nil {
				b.log.Error("Receive message failed",
					zap.String("op", op),
					zap.Error(err))
				continue
			}

			for _, entry := range entries[0].Messages {
				userID, ok := entry.Values["user_id"].(string)
				if !ok {
					b.log.Error("Invalid message type",
						zap.String("op", op),
						zap.Error(err))
					continue
				}

				msgChan <- userID

				b.channel.XAck(ctx, stream, group, entry.ID)
			}
		}
	}()

	return msgChan, nil
}
