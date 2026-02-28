package users

import (
	"net/http"

	"gateway/internal/services"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/gin-gonic/gin"
)

const redisTTL = 3600 * 24 * 30

func (us *UsersService) Register(c *gin.Context) {
	req := struct {
		Name     string `json:"name"     validate:"required"`
		Email    string `json:"email"    validate:"required,email"`
		Role     string `json:"role"     validate:"oneof=admin user guest dev"`
		Password string `json:"password" validate:"required,min=8"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := us.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	res, err := services.Execute(us.cb, func() (*pb.RegRes, error) {
		return us.client.RegUser(c.Request.Context(), &pb.RegReq{
			Name:     req.Name,
			Email:    req.Email,
			Role:     req.Role,
			Password: req.Password,
		})
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"session_key",
		res.SessionKey,
		redisTTL, "/", "localhost",
		false, true,
	)

	c.JSON(http.StatusOK, gin.H{"token": res.Token, "user_id": res.UserId})
}

func (us *UsersService) Login(c *gin.Context) {
	req := struct {
		Name     string `json:"name"     validate:"required"`
		Email    string `json:"email"    validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := us.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	res, err := services.Execute(us.cb, func() (*pb.LogRes, error) {
		return us.client.LogUser(c.Request.Context(), &pb.LogReq{
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
		})
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"session_key",
		res.SessionKey,
		redisTTL, "/", "localhost",
		false, true,
	)

	c.JSON(http.StatusOK, gin.H{"token": res.Token, "user_id": res.UserId})
}

func (us *UsersService) DeleteUser(c *gin.Context) {
	req := struct {
		DelUserID  string `validate:"required,uuid"`
		sessionKey string
		token      string
	}{}

	req.DelUserID = c.Param("del_user_id")
	req.sessionKey = c.GetString("session_key")
	req.token = c.GetString("token")

	if err := us.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	if _, err := services.Execute(us.cb, func() (*pb.DelRes, error) {
		return us.client.DelUser(c.Request.Context(), &pb.DelReq{
			DelUserId:  req.DelUserID,
			SessionKey: req.sessionKey,
			Token:      req.token,
		})
	}); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
