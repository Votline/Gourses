package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"users/internal/broker"
	"users/internal/db"
	gc "users/internal/gracefulshutdown"
	"users/internal/rdb"
	"users/internal/security"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type usersserver struct {
	log *zap.Logger
	db  *db.DB
	rdb *rdb.RDB
	brk *broker.Broker
	pb.UnimplementedUsersServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	defer log.Sync()

	lis, err := net.Listen("tcp", ":"+os.Getenv("USERS_PORT"))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := db.NewDB(log)
	if err != nil {
		log.Fatal("Failed to create database", zap.Error(err))
	}
	defer db.Close(ctx)
	log.Info("Connected to database")

	rdb, err := rdb.NewRDB(log)
	if err != nil {
		log.Fatal("Failed to create redis", zap.Error(err))
	}
	defer rdb.Close(ctx)
	log.Info("Connected to redis")

	broker, err := broker.NewBroker(log)
	if err != nil {
		log.Fatal("Failed to create broker", zap.Error(err))
	}
	defer broker.Close(ctx)
	log.Info("Connected to broker")

	u := usersserver{log: log, db: db, rdb: rdb, brk: broker}
	srv := grpc.NewServer()
	pb.RegisterUsersServiceServer(srv, &u)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	gracefulShutdown(&u, srv, log)
}

func gracefulShutdown(u *usersserver, srv *grpc.Server, log *zap.Logger) {
	log.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down gRPC server")
	if err := gc.Shutdown(
		func() error { srv.GracefulStop(); return nil }, ctx); err != nil {
		log.Error("Failed to shutdown server", zap.Error(err))
	}

	log.Info("Shutting down postgres")
	if err := u.db.Close(ctx); err != nil {
		log.Error("Failed to shutdown postgres", zap.Error(err))
	}

	log.Info("Shutting down redis")
	if err := u.rdb.Close(ctx); err != nil {
		log.Error("Failed to shutdown redis", zap.Error(err))
	}

	log.Info("Shutting down broker")
	if err := u.brk.Close(ctx); err != nil {
		log.Error("Failed to shutdown broker", zap.Error(err))
	}
}

func (u *usersserver) RegUser(ctx context.Context, req *pb.RegReq) (*pb.RegRes, error) {
	const op = "usersserver.RegUser"

	name := req.GetName()
	email := req.GetEmail()
	role := req.GetRole()
	pswd := req.GetPassword()

	id := uuid.New().String()
	hashPswd, err := security.Hash(pswd)
	if err != nil {
		return nil, fmt.Errorf("%s: hash password: %w", op, err)
	}

	token, err := security.GenerateToken(id, role)
	if err != nil {
		return nil, fmt.Errorf("%s: generate token: %w", op, err)
	}

	sk, err := u.rdb.NewSession(id, role)
	if err != nil {
		return nil, fmt.Errorf("%s: create session: %w", op, err)
	}

	if err := u.db.RegUser(id, name+email, role, hashPswd); err != nil {
		u.log.Error(op, zap.Error(err))
		if err := u.rdb.Delete(sk); err != nil {
			u.log.Error("delete session",
				zap.String("op", op),
				zap.Error(err))
		}
		return nil, fmt.Errorf("%s: add to db: %w", op, err)
	}

	return &pb.RegRes{Token: token, SessionKey: sk, UserId: id}, nil
}

func (u *usersserver) LogUser(ctx context.Context, req *pb.LogReq) (res *pb.LogRes, err error) {
	const op = "usersserver.LogUser"

	name := req.GetName()
	email := req.GetEmail()
	pswd := req.GetPassword()

	ui, err := u.db.LogUser(name + email)
	if err != nil {
		return nil, fmt.Errorf("%s: log user: %w", op, err)
	}

	token, err := security.GenerateToken(ui.ID, ui.Role)
	if err != nil {
		return nil, fmt.Errorf("%s: generate token: %w", op, err)
	}

	sk, err := u.rdb.NewSession(ui.ID, ui.Role)
	if err != nil {
		return nil, fmt.Errorf("%s: create session: %w", op, err)
	}

	if err = security.Check(pswd, ui.Pswd); err != nil {
		return nil, fmt.Errorf("%s: check password: %w", op, err)
	}

	return &pb.LogRes{Token: token, SessionKey: sk, UserId: ui.ID}, nil
}

func (u *usersserver) UpdateUser(ctx context.Context, req *pb.UpdateReq) (*pb.UpdateRes, error) {
	const op = "usersserver.UpdateUser"

	id := req.GetUserId()
	role := req.GetUserRole()
	sk := req.GetSessionKey()
	name := req.GetNewName()
	email := req.GetNewEmail()
	newRole := req.GetNewRole()
	pswd := req.GetNewPassword()

	if err := u.rdb.Validate(id, role, sk); err != nil {
		return nil, fmt.Errorf("%s: validate session: %w", op, err)
	}

	hashedPswd, err := security.Hash(pswd)
	if err != nil {
		return nil, fmt.Errorf("%s: hash password: %w", op, err)
	}

	newToken, err := security.GenerateToken(id, newRole)
	if err != nil {
		return nil, fmt.Errorf("%s: generate token: %w", op, err)
	}

	newSk := sk
	if role != newRole {
		if err := u.rdb.Delete(sk); err != nil {
			return nil, fmt.Errorf("%s: delete session: %w", op, err)
		}
		newSk, err = u.rdb.NewSession(id, newRole)
		if err != nil {
			return nil, fmt.Errorf("%s: create session: %w", op, err)
		}
	}

	if err := u.db.UpdateUser(id, name+email, newRole, hashedPswd); err != nil {
		u.log.Error(op, zap.Error(err))
		if err := u.rdb.Delete(sk); err != nil {
			u.log.Error("delete session",
				zap.String("op", op),
				zap.Error(err))
		}
		return nil, fmt.Errorf("%s: update user: %w", op, err)
	}

	return &pb.UpdateRes{Token: newToken, SessionKey: newSk}, nil
}

func (u *usersserver) DelUser(ctx context.Context, req *pb.DelReq) (*pb.DelRes, error) {
	const op = "usersserver.DelUser"

	id := req.GetUserId()
	role := req.GetUserRole()
	sk := req.GetSessionKey()
	delUserID := req.GetDelUserId()

	if err := u.rdb.Validate(id, role, sk); err != nil {
		return nil, fmt.Errorf("%s: validate session: %w", op, err)
	}

	if err := u.rdb.Delete(sk); err != nil {
		return nil, fmt.Errorf("%s: delete session: %w", op, err)
	}

	if err := u.db.DelUser(id, role, delUserID); err != nil {
		u.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: delete user: %w", op, err)
	}

	if err := u.brk.Publish("users:delete", id); err != nil {
		u.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: publish message: %w", op, err)
	}

	return &pb.DelRes{}, nil
}

func (u *usersserver) ValidateUser(ctx context.Context, req *pb.ValidateReq) (*pb.ValidateRes, error) {
	const op = "usersserver.ValidateUser"

	token := req.GetToken()
	sk := req.GetSessionKey()

	userInfo, err := security.ExtractUnverifiedClaims(token)
	if err != nil {
		return nil, fmt.Errorf("%s: extract user info: %w", op, err)
	}

	if err := u.rdb.Validate(userInfo.ID, userInfo.Role, sk); err != nil {
		return nil, fmt.Errorf("%s: validate session: %w", op, err)
	}

	userInfo, err = u.rdb.Extract(sk)
	if err != nil {
		return nil, fmt.Errorf("%s: extract user info: %w", op, err)
	}

	newToken, err := security.GenerateToken(userInfo.ID, userInfo.Role)
	if err != nil {
		return nil, fmt.Errorf("%s: generate token: %w", op, err)
	}

	return &pb.ValidateRes{Token: newToken}, nil
}
