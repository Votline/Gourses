package courses

import (
	"net/http"
	"strconv"

	"gateway/internal/services"

	pb "github.com/Votline/Gourses/protos/generated-courses"
	"github.com/gin-gonic/gin"
)

func (cs *CoursesService) NewCourse(c *gin.Context) {
	const op = "courses.NewCourse"

	req := struct {
		UserID   string `validate:"required,uuid"`
		UserRole string `validate:"required,eq=admin"`
		Name     string `json:"name"  validate:"required"`
		Desc     string `json:"desc"  validate:"required"`
		Price    string `json:"price" validate:"required"`
	}{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	if _, err := strconv.Atoi(req.Price); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid price"})
		return
	}

	req.UserID = c.GetString("user_id")
	req.UserRole = c.GetString("user_role")

	if err := cs.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	res, err := services.Execute(cs.cb, func() (*pb.NewCourseRes, error) {
		return cs.client.NewCourse(c.Request.Context(), &pb.NewCourseReq{
			UserId:     req.UserID,
			UserRole:   req.UserRole,
			Name:       req.Name,
			Desciption: req.Desc,
			Price:      req.Price,
		})
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"course_id": res.CourseId})
}
