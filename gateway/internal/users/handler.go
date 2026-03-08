package users

import (
	"context"
	"net/http"

	"gateway/internal/services"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const redisTTL = 3600 * 24 * 30

func (us *UsersService) Register(c *gin.Context) {
	req := struct {
		Name     string `json:"name"     validate:"required"`
		Email    string `json:"email"    validate:"required,email"`
		Role     string `json:"role"     validate:"oneof=admin teacher user guest dev"`
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

func (us *UsersService) UpdateUser(c *gin.Context) {
	req := struct {
		UserID      string `validate:"required,uuid"`
		SessionKey  string `validate:"required,uuid"`
		UserRole    string `validate:"required,oneof=admin teacher user guest dev"`
		NewRole     string `json:"new_role" validate:"required,oneof=admin teacher user guest dev"`
		NewName     string `json:"new_name" validate:"required"`
		NewEmail    string `json:"new_email" validate:"required,email"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}{}

	req.UserID = c.GetString("user_id")
	req.UserRole = c.GetString("user_role")
	req.SessionKey = c.GetString("session_key")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	us.log.Info("user update request",
		zap.String("user_id", req.UserID),
		zap.String("user_role", req.UserRole),
		zap.String("session_key", req.SessionKey),
		zap.String("new_role", req.NewRole),
		zap.String("new_name", req.NewName),
		zap.String("new_email", req.NewEmail),
		zap.String("new_password", req.NewPassword),
	)

	if err := us.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	res, err := services.Execute(us.cb, func() (*pb.UpdateRes, error) {
		return us.client.UpdateUser(c.Request.Context(), &pb.UpdateReq{
			UserId:      req.UserID,
			UserRole:    req.UserRole,
			SessionKey:  req.SessionKey,
			NewName:     req.NewName,
			NewEmail:    req.NewEmail,
			NewRole:     req.NewRole,
			NewPassword: req.NewPassword,
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

	c.JSON(http.StatusOK, gin.H{"token": res.Token, "user_id": req.UserID})
}

func (us *UsersService) DeleteUser(c *gin.Context) {
	req := struct {
		DelUserID  string `validate:"required,uuid"`
		SessionKey string `validate:"required,uuid"`
		UserID     string `validate:"required,uuid"`
		UserRole   string `validate:"required,oneof=admin teacher user guest dev"`
	}{}

	req.DelUserID = c.Param("del_user_id")
	req.SessionKey = c.GetString("session_key")
	req.UserID = c.GetString("user_id")
	req.UserRole = c.GetString("user_role")

	if err := us.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	if _, err := services.Execute(us.cb, func() (*pb.DelRes, error) {
		return us.client.DelUser(c.Request.Context(), &pb.DelReq{
			DelUserId:  req.DelUserID,
			SessionKey: req.SessionKey,
			UserId:     req.UserID,
			UserRole:   req.UserRole,
		})
	}); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (us *UsersService) Validate(ctx context.Context, tokenStr, sessionKey string) (*pb.ValidateRes, error) {
	res, err := services.Execute(us.cb, func() (*pb.ValidateRes, error) {
		return us.client.ValidateUser(ctx, &pb.ValidateReq{
			Token: tokenStr,
		})
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
