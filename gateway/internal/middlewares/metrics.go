package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func (m *Mdwr) Metrics(newTimer func(name, operation string) *prometheus.Timer, incr func(name string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Split(strings.Trim(c.Request.URL.String(), "/"), "/")
		// [api, serviceName, operation]

		if len(parts) < 3 {
			c.Next()
			return
		}

		serviceName := parts[1]
		operation := parts[2]

		timer := newTimer(serviceName, operation)
		defer timer.ObserveDuration()

		incr(serviceName)

		c.Next()
	}
}
