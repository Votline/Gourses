package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const jwtLength = 100

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header must be 'Bearer {token}'"})
			return
		}

		tokenStr := parts[1]
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Token not found"})
			return
		}
		if len(tokenStr) < jwtLength {
			fmt.Printf("DEBUG: Received header: [%s]\n==Len?:%v",
				c.GetHeader("Authorization"), len(tokenStr) == jwtLength)

			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid token"})
			return
		}

		c.Set("token", tokenStr)
		c.Next()
	}
}
