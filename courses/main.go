package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"courses/internal/db"

	pb "github.com/Votline/Gourses/protos/generated-courses"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type coursesservice struct {
	log *zap.Logger
	db  *db.DB
	pb.UnimplementedCoursesServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	defer log.Sync()

	lis, err := net.Listen("tcp", ":"+os.Getenv("COURSES_PORT"))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	db, err := db.NewDB(log)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}
	defer db.Close()
	log.Info("Connected to database")

	c := coursesservice{log: log, db: db}
	srv := grpc.NewServer()
	pb.RegisterCoursesServiceServer(srv, &c)
	if err := srv.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
	defer srv.Stop()
}

func (c *coursesservice) NewCourse(ctx context.Context, req *pb.NewCourseReq) (*pb.NewCourseRes, error) {
	const op = "courses.NewCourse"

	userID := req.GetUserId()
	userRole := req.GetUserRole()
	name := req.GetName()
	desc := req.GetDescription()
	price := req.GetPrice()

	id := uuid.NewString()

	if err := c.db.NewCourse(id, name, desc, price, userID, userRole); err != nil {
		c.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: new course: %w", op, err)
	}

	return &pb.NewCourseRes{CourseId: id}, nil
}

func (c *coursesservice) GetCourse(ctx context.Context, req *pb.GetCourseReq) (*pb.GetCourseRes, error) {
	const op = "courses.GetCourses"

	coursesID := req.GetCourseId()

	courseInfo, err := c.db.GetCourse(coursesID)
	if err != nil {
		c.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: get courses: %w", op, err)
	}

	return &pb.GetCourseRes{
		CourseId:   coursesID,
		Name:       courseInfo.Name,
		Desciption: courseInfo.Desc,
		Price:      courseInfo.Price,
	}, nil
}

func (c *coursesservice) UpdateCourse(ctx context.Context, req *pb.UpdateCourseReq) (*pb.UpdateCourseRes, error) {
	const op = "courses.UpdateCourse"

	userID := req.GetUserId()
	userRole := req.GetUserRole()
	courseID := req.GetCourseId()
	name := req.GetNewName()
	desc := req.GetNewDescription()
	price := req.GetNewPrice()

	if err := c.db.UpdateCourse(userID, userRole, courseID, name, desc, price); err != nil {
		c.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: update course: %w", op, err)
	}

	return &pb.UpdateCourseRes{}, nil
}

func (c *coursesservice) DeleteCourse(ctx context.Context, req *pb.DeleteCourseReq) (*pb.DeleteCourseRes, error) {
	const op = "courses.DeleteCourse"

	courseID := req.GetCourseId()
	userID := req.GetUserId()
	userRole := req.GetUserRole()

	if err := c.db.DeleteCourse(courseID, userID, userRole); err != nil {
		c.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: delete course: %w", op, err)
	}

	return &pb.DeleteCourseRes{}, nil
}
