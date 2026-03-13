package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"courses/internal/broker"
	"courses/internal/db"
	gc "courses/internal/gracefulshutdown"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := db.NewDB(log)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}
	defer db.Close(ctx)
	log.Info("Connected to database")

	broker, err := broker.NewBroker(log)
	if err != nil {
		log.Fatal("failed to create broker", zap.Error(err))
	}
	defer broker.Close(ctx)
	log.Info("Connected to broker")

	c := coursesservice{log: log, db: db}

	go c.listenDelete(ctx, broker)

	srv := grpc.NewServer()
	pb.RegisterCoursesServiceServer(srv, &c)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatal("failed to serve", zap.Error(err))
		}
		defer srv.Stop()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	gracefulShutdown(&c, srv, broker, log)
}

func gracefulShutdown(c *coursesservice, srv *grpc.Server, broker *broker.Broker, log *zap.Logger) {
	const op = "courses.gracefulShutdown"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down gRPC server")
	if err := gc.Shutdown(
		func() error { srv.GracefulStop(); return nil }, ctx); err != nil {
		log.Error(op, zap.Error(err))
	}

	log.Info("Shutting down postgres")
	if err := c.db.Close(ctx); err != nil {
		log.Error(op, zap.Error(err))
	}

	log.Info("Shutting down broker")
	if err := broker.Close(ctx); err != nil {
		log.Error(op, zap.Error(err))
	}

	log.Info("Shutting down")
}

func (c *coursesservice) NewCourse(ctx context.Context, req *pb.NewCourseReq) (*pb.NewCourseRes, error) {
	const op = "courses.NewCourse"

	userID := req.GetUserId()
	name := req.GetName()
	desc := req.GetDescription()
	price := req.GetPrice()

	id := uuid.NewString()

	if err := c.db.NewCourse(id, name, desc, price, userID); err != nil {
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

func (c *coursesservice) listenDelete(ctx context.Context, broker *broker.Broker) {
	const op = "courses.listenDelete"

	sub := broker.Subscribe(ctx, "users:delete")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub:
			if !ok {
				c.log.Info("Broker closed")
				return
			}

			c.log.Debug("Received message",
				zap.String("op", op),
				zap.String("msg", msg))

			err := c.deleteCourseByID(msg)
			if err != nil {
				c.log.Error(op, zap.Error(err))
			}
		}
	}
}

func (c *coursesservice) deleteCourseByID(id string) error {
	const op = "courses.deleteCourseByID"

	if err := c.db.DeleteCourseByID(id); err != nil {
		c.log.Error(op, zap.Error(err))
		return fmt.Errorf("%s: delete course: %w", op, err)
	}

	return nil
}
