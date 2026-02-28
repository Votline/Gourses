package middlewares

import (
	"strings"

	"gateway/internal/services"

	"github.com/gin-gonic/gin"
)

func Metrics(svc services.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Split(strings.Trim(c.Request.URL.String(), "/"), "/")
		// [api, serviceName, operation]

		if len(parts) < 3 {
			c.Next()
			return
		}

		serviceName := parts[1]
		operation := parts[2]

		timer := svc.NewTimer(serviceName, operation)
		defer timer.ObserveDuration()

		svc.IncrCounter(operation)

		c.Next()
	}
}
