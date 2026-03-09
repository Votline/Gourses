package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	exp     = time.Minute
	maxRate = 50
)

func (m *Mdwr) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Copy().RemoteIP()
		key := "rl:" + ip

		count, err := m.rdb.Incr(m.ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "redis incr: " + err.Error()})
			return
		}

		if count == 1 {
			m.rdb.Expire(m.ctx, key, exp)
		}

		if count > maxRate {
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
