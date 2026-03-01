package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const sessionKeyLength = 36

func SessionKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, err := ExtractSessionKey(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": err.Error()})
			return
		}

		c.Set("session_key", sessionKey)
		c.Next()
	}
}

func ExtractSessionKey(c *gin.Context) (string, error) {
	sessionKey, err := c.Cookie("session_key")
	if err != nil {
		return "", err
	}

	if len(sessionKey) != sessionKeyLength {
		return "", errors.New("Invalid session key")
	}

	return sessionKey, nil
}
