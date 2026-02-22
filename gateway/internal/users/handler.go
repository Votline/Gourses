package users

import (
	"net/http"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/gin-gonic/gin"
)

func (us *UserService) Register(c *gin.Context) {
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

	res, err := us.client.RegUser(c.Request.Context(), &pb.RegReq{
		Name:     req.Name,
		Email:    req.Email,
		Role:     req.Role,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": res.Token})
}
