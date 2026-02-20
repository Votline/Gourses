package routers

import (
	"fmt"

	"gourses/internal/services"
	"gourses/internal/users"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	r := gin.Default()

	initServices(r)

	return r
}

func initServices(r *gin.Engine) {
	svcs := [1]services.Service{users.New()}
	for _, svc := range svcs {
		path := "/api/" + svc.GetName()
		group := r.Group(path)

		svc.RegisterRoutes(group)

		fmt.Println("\n\n\nRegistered service: " + path)
	}
}
