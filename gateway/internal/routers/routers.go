package routers

import (
	"gateway/internal/services"
	"gateway/internal/users"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Init(log *zap.Logger) *gin.Engine {
	r := gin.Default()

	initServices(r, log)

	return r
}

func initServices(r *gin.Engine, log *zap.Logger) {
	svcs := [1]services.Service{users.New(log)}
	for _, svc := range svcs {
		path := "/api/" + svc.GetName()
		group := r.Group(path)

		svc.RegisterRoutes(group)
	}
}
