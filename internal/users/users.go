package users

import (
	"gourses/internal/services"

	"github.com/gin-gonic/gin"
)

type UserService struct {
	name string
}

func New() services.Service {
	return &UserService{name: "users"}
}

func (us *UserService) GetName() string {
	return us.name
}

func (us *UserService) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/reg", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})
}
