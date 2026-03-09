package services

import (
	"errors"
	"time"

	"gateway/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	GetName() string
	RegisterRoutes(r *gin.RouterGroup, mdwr *middlewares.Mdwr)
	IncrCounter(name string)
	NewTimer(name, method string) *prometheus.Timer
}

func Execute[T any](cb *gobreaker.CircuitBreaker[any], fn func() (T, error)) (T, error) {
	var zero T

	resCb, err := cb.Execute(func() (any, error) {
		return rpcRetry(func() (T, error) {
			return fn()
		})
	})

	res, ok := resCb.(T)
	if !ok {
		return zero, errors.New("invalid type")
	}
	return res, err
}

func rpcRetry[T any](fn func() (T, error)) (T, error) {
	var zero T

	for i := range 5 {
		res, err := fn()
		if err == nil {
			return res, nil
		}

		if !shouldRetry(err) {
			return zero, err
		}

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return zero, errors.New("max retries exceeded")
}

func shouldRetry(err error) bool {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case
			codes.Canceled,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Aborted,
			codes.Unavailable,
			codes.DataLoss:

			return true
		case
			codes.InvalidArgument,
			codes.NotFound,
			codes.AlreadyExists,
			codes.PermissionDenied,
			codes.FailedPrecondition,
			codes.OutOfRange,
			codes.Unimplemented,
			codes.Internal,
			codes.Unauthenticated:

			return false
		}
	}

	return false
}
