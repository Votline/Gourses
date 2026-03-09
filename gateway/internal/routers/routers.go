package routers

import (
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

func Init(log *zap.Logger) *gin.Engine {
	r := gin.Default()

	initServices(r, log)
	r.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	return r
}

func initServices(r *gin.Engine, log *zap.Logger) {
	us := users.New(log, resTime).(*users.UsersService)

	mdwr, err := middlewares.NewMdwr(us.Validate)
	if err != nil {
		log.Fatal("NewMdwr", zap.Error(err))
	}

	svcs := [2]services.Service{
		us,
		courses.New(log, resTime),
	}

	r.Use(mdwr.RateLimit())

	for _, svc := range svcs {
		path := "/api/" + svc.GetName()
		group := r.Group(path)

		svc.RegisterRoutes(group, mdwr)
	}
}
