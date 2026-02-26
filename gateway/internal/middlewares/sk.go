package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const sessionKeyLength = 36

func SessionKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, err := c.Cookie("session_key")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": err.Error()})
			return
		}

		if len(sessionKey) != sessionKeyLength {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid session key"})
			return
		}

		c.Set("session_key", sessionKey)
		c.Next()
	}
}
