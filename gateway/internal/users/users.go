package users

import (
	"os"

	"gateway/internal/services"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UserService struct {
	name   string
	log    *zap.Logger
	val    *validator.Validate
	conn   *grpc.ClientConn
	client pb.UsersServiceClient
}

func New(log *zap.Logger) services.Service {
	log.Debug("Creating users service", zap.String("address", os.Getenv("USERS_HOST")+":"+os.Getenv("USERS_PORT")))
	conn, err := grpc.NewClient(
		os.Getenv("USERS_HOST")+":"+os.Getenv("USERS_PORT"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed connect to users service", zap.Error(err))
	}
	return &UserService{
		name:   "users",
		log:    log,
		val:    validator.New(),
		conn:   conn,
		client: pb.NewUsersServiceClient(conn),
	}
}

func (us *UserService) GetName() string {
	return us.name
}

func (us *UserService) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/reg", func(c *gin.Context) {
		us.Register(c)
	})
}
