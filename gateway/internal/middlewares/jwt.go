package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtLength = 100

func (m *Mdwr) JWTMiddleware() gin.HandlerFunc {
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

		userInfo, err := m.checkJWT(tokenStr, c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		c.Set("user_id", userInfo.ID)
		c.Set("user_role", userInfo.Role)
		c.Next()
	}
}

func (m *Mdwr) checkJWT(tokenStr string, c *gin.Context) (UserInfo, error) {
	const op = "middlewares.CheckJWT"

	claims, err := extractJWTData(tokenStr)
	if err != nil {
		return UserInfo{}, fmt.Errorf("%s: extract jwt: %w", op, err)
	}

	if !claims.token.Valid {
		sk, err := ExtractSessionKey(c)
		if err != nil {
			return UserInfo{}, fmt.Errorf("%s: extract session key %w", op, err)
		}
		res, err := m.validate(c.Request.Context(), tokenStr, sk)
		if err != nil {
			return UserInfo{}, fmt.Errorf("%s: rpc validate %w", op, err)
		}
		if res.Token == "" {
			return UserInfo{}, fmt.Errorf("%s: Invalid token", op)
		}
	}

	return claims, nil
}

func extractJWTData(tokenStr string) (UserInfo, error) {
	claims := UserInfo{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return UserInfo{}, err
	}
	claims.token = token
	return claims, nil
}
