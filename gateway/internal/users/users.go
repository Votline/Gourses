package users

import (
	"os"

	"gateway/internal/cbreaker"
	"gateway/internal/middlewares"
	"gateway/internal/services"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UsersService struct {
	name           string
	log            *zap.Logger
	val            *validator.Validate
	conn           *grpc.ClientConn
	client         pb.UsersServiceClient
	cb             *gobreaker.CircuitBreaker[any]
	metricsCounter *prometheus.CounterVec
	metricsHist    *prometheus.HistogramVec
}

func New(log *zap.Logger, resTime *prometheus.HistogramVec) services.Service {
	log.Debug("Creating users service", zap.String("address", os.Getenv("USERS_HOST")+":"+os.Getenv("USERS_PORT")))
	conn, err := grpc.NewClient(
		os.Getenv("USERS_HOST")+":"+os.Getenv("USERS_PORT"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed connect to users service", zap.Error(err))
	}
	return &UsersService{
		name:        "users",
		log:         log,
		val:         validator.New(),
		conn:        conn,
		client:      pb.NewUsersServiceClient(conn),
		cb:          cbreaker.NewCircuitBreaker("users", log),
		metricsHist: resTime,
		metricsCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_users_counter",
				Help: "Counter for gateway users service",
			},
			[]string{"operation"},
		),
	}
}

func (us *UsersService) GetName() string {
	return us.name
}

func (us *UsersService) RegisterRoutes(r *gin.RouterGroup) {
	r.Use(middlewares.Metrics(us))
	r.POST("/reg", us.Register)
	r.POST("/log", us.Login)

	verifyGroup := r.Group("")
	verifyGroup.Use(middlewares.JWTMiddleware(us.Validate))
	verifyGroup.Use(middlewares.SessionKeyMiddleware())
	{
		verifyGroup.DELETE("/del/:del_user_id", us.DeleteUser)
	}
}

func (us *UsersService) IncrCounter(name string) {
	us.metricsCounter.WithLabelValues(name).Inc()
}

func (us *UsersService) NewTimer(name, operation string) *prometheus.Timer {
	return prometheus.NewTimer(prometheus.ObserverFunc(func(t float64) {
		us.metricsHist.WithLabelValues(name, operation).Observe(t)
	}))
}
