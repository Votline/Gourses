package users

import (
	"gateway/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserService struct {
	name string
	val  *validator.Validate
}

func New() services.Service {
	return &UserService{name: "users", val: validator.New()}
}

func (us *UserService) GetName() string {
	return us.name
}

func (us *UserService) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/reg", func(c *gin.Context) {
		us.Register(c)
	})
}
