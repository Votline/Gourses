package cbreaker

import (
	"time"

	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
)

func NewCircuitBreaker(name string, log *zap.Logger) *gobreaker.CircuitBreaker[any] {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 10,
		Interval:    time.Minute,
		Timeout:     2 * time.Minute,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Info("CircuitBreaker changed",
				zap.String("service", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}
	return gobreaker.NewCircuitBreaker[any](st)
}
