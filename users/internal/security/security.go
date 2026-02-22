package security

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Check(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func GenerateToken(id, role string) (string, error) {
	claims := UserInfo{
		ID:   id,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ExtractJWTData(tokenStr string) (UserInfo, error) {
	claims := UserInfo{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return os.Getenv("JWT_SECRET"), nil
	})
	if err != nil {
		return UserInfo{}, err
	}

	if !token.Valid {
		return UserInfo{}, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
