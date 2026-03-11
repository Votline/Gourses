package routers

import (
	"context"
	"net/http"
	"os"
	"sync"

	"gateway/internal/courses"
	"gateway/internal/middlewares"
	"gateway/internal/services"
	"gateway/internal/users"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var resTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "seconds_per_operation",
	Help:    "Time spent processing requests",
	Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
}, []string{"service", "operation"})

type Server struct {
	Srv  *http.Server
	svcs []services.Service
}

func Init(log *zap.Logger) *Server {
	r := gin.Default()

	svcs := initServices(r, log)
	r.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	srv := &http.Server{
		Addr:    ":" + os.Getenv("API_PORT"),
		Handler: r,
	}

	return &Server{Srv: srv, svcs: svcs}
}

func initServices(r *gin.Engine, log *zap.Logger) []services.Service {
	us := users.New(log, resTime).(*users.UsersService)

	mdwr, err := middlewares.NewMdwr(us.Validate)
	if err != nil {
		log.Fatal("NewMdwr", zap.Error(err))
	}

	svcs := []services.Service{
		us,
		courses.New(log, resTime),
	}

	r.Use(mdwr.RateLimit())

	for _, svc := range svcs {
		path := "/api/" + svc.GetName()
		group := r.Group(path)

		svc.RegisterRoutes(group, mdwr)
	}

	return svcs
}

func (s *Server) ShutdownServices(ctx context.Context, log *zap.Logger) error {
	var wg sync.WaitGroup
	done := make(chan struct{})

	for _, svc := range s.svcs {
		log.Info("Shutting down " + svc.GetName() + " service")
		wg.Go(func() {
			svc.Close(ctx)
		})
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
