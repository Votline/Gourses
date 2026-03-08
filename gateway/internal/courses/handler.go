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
		UserID string `validate:"required,uuid"`
		Name   string `json:"name"  validate:"required"`
		Desc   string `json:"description"  validate:"required"`
		Price  string `json:"price" validate:"required"`
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

	if err := cs.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	res, err := services.Execute(cs.cb, func() (*pb.NewCourseRes, error) {
		return cs.client.NewCourse(c.Request.Context(), &pb.NewCourseReq{
			UserId:      req.UserID,
			Name:        req.Name,
			Description: req.Desc,
			Price:       req.Price,
		})
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"course_id": res.CourseId})
}

func (cs *CoursesService) GetCourse(c *gin.Context) {
	const op = "courses.GetCourse"

	req := struct {
		CourseID string `validate:"required,uuid"`
	}{}
	req.CourseID = c.Param("course_id")

	if err := cs.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	res, err := services.Execute(cs.cb, func() (*pb.GetCourseRes, error) {
		return cs.client.GetCourse(c.Request.Context(), &pb.GetCourseReq{
			CourseId: req.CourseID,
		})
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"course_id":   res.CourseId,
		"name":        res.Name,
		"description": res.Desciption,
		"price":       res.Price,
	})
}

func (cs *CoursesService) DeleteCourse(c *gin.Context) {
	const op = "courses.DeleteCourse"

	req := struct {
		CourseID string `validate:"required,uuid"`
		UserID   string `validate:"required,uuid"`
	}{}
	req.CourseID = c.Param("course_id")
	req.UserID = c.GetString("user_id")

	if err := cs.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	if _, err := services.Execute(cs.cb, func() (*pb.DeleteCourseRes, error) {
		return cs.client.DeleteCourse(c.Request.Context(), &pb.DeleteCourseReq{
			CourseId: req.CourseID,
			UserId:   req.UserID,
		})
	}); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (cs *CoursesService) UpdateCourse(c *gin.Context) {
	const op = "courses.UpdateCourse"

	req := struct {
		CourseID string `validate:"required,uuid"`
		UserID   string `validate:"required,uuid"`
		UserRole string `validate:"required,oneof=admin teacher"`
		Name     string `json:"new_name"  validate:"required"`
		Desc     string `json:"new_description"  validate:"required"`
		Price    string `json:"new_price" validate:"required"`
	}{}

	req.CourseID = c.Param("course_id")
	req.UserID = c.GetString("user_id")
	req.UserRole = c.GetString("user_role")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	if err := cs.val.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	if _, err := strconv.Atoi(req.Price); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid price"})
		return
	}

	if _, err := services.Execute(cs.cb, func() (*pb.UpdateCourseRes, error) {
		return cs.client.UpdateCourse(c.Request.Context(), &pb.UpdateCourseReq{
			UserId:         req.UserID,
			UserRole:       req.UserRole,
			CourseId:       req.CourseID,
			NewName:        req.Name,
			NewDescription: req.Desc,
			NewPrice:       req.Price,
		})
	}); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
