package services

import "github.com/gin-gonic/gin"

type Service interface {
	GetName() string
	RegisterRoutes(r *gin.RouterGroup)
}
