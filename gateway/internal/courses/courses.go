package courses

import (
	"context"
	"os"

	"gateway/internal/cbreaker"
	gc "gateway/internal/gracefulshutdown"
	"gateway/internal/middlewares"
	"gateway/internal/services"

	pb "github.com/Votline/Gourses/protos/generated-courses"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type CoursesService struct {
	name           string
	log            *zap.Logger
	val            *validator.Validate
	conn           *grpc.ClientConn
	client         pb.CoursesServiceClient
	cb             *gobreaker.CircuitBreaker[any]
	metricsCounter *prometheus.CounterVec
	metricsHist    *prometheus.HistogramVec
}

func New(log *zap.Logger, resTime *prometheus.HistogramVec) services.Service {
	log.Debug("Creating courses service",
		zap.String("address",
			os.Getenv("COURSES_HOST")+":"+os.Getenv("COURSES_PORT")),
	)

	conn, err := grpc.NewClient(
		os.Getenv("COURSES_HOST")+":"+os.Getenv("COURSES_PORT"),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal("Failed connect to courses service", zap.Error(err))
	}

	return &CoursesService{
		name:        "courses",
		log:         log,
		val:         validator.New(),
		conn:        conn,
		client:      pb.NewCoursesServiceClient(conn),
		cb:          cbreaker.NewCircuitBreaker("courses", log),
		metricsHist: resTime,
		metricsCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_courses_counter",
				Help: "Counter for gateway courses service",
			},
			[]string{"operation"},
		),
	}
}

func (cs *CoursesService) GetName() string {
	return cs.name
}

func (cs *CoursesService) RegisterRoutes(r *gin.RouterGroup, mdwr *middlewares.Mdwr) {
	r.Use(mdwr.Metrics(cs.NewTimer, cs.IncrCounter))
	r.Use(mdwr.JWTMiddleware())

	r.POST("/new", cs.NewCourse)
	r.GET("/get/:course_id", cs.GetCourse)
	r.DELETE("/delete/:course_id", cs.DeleteCourse)
	r.PUT("/update/:course_id", cs.UpdateCourse)
}

func (cs *CoursesService) IncrCounter(name string) {
	cs.metricsCounter.WithLabelValues(name).Inc()
}

func (cs *CoursesService) NewTimer(name, operation string) *prometheus.Timer {
	return prometheus.NewTimer(prometheus.ObserverFunc(func(t float64) {
		cs.metricsHist.WithLabelValues(name, operation).Observe(t)
	}))
}

func (cs *CoursesService) Close(ctx context.Context) error {
	return gc.Shutdown(cs.conn.Close, ctx)
}
