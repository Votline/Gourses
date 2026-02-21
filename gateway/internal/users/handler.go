package users

import (
	"github.com/gin-gonic/gin"
)

func (us *UserService) Register(c *gin.Context) {
	req := struct {
		Name     string `json:"name"     validate:"required"`
		Email    string `json:"email"    validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if us.val.Struct(req) != nil {
		c.JSON(400, gin.H{"error": "invalid request", "name": req.Name, "email": req.Email, "password": req.Password})
		return
	}

	c.JSON(200, gin.H{"name": req.Name, "email": req.Email, "password": req.Password})
}
