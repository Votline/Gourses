package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"users/internal/db"
	"users/internal/security"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type usersserver struct {
	log *zap.Logger
	db  *db.DB
	pb.UnimplementedUsersServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	defer log.Sync()

	lis, err := net.Listen("tcp", ":"+os.Getenv("USERS_PORT"))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	db, err := db.NewDB(log)
	if err != nil {
		log.Fatal("Failed to create database", zap.Error(err))
	}
	defer db.Close()

	u := usersserver{log: log, db: db}
	srv := grpc.NewServer()
	pb.RegisterUsersServiceServer(srv, &u)
	if err := srv.Serve(lis); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
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

	if err := u.db.RegUser(id, name+email, role, hashPswd); err != nil {
		u.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: add to db: %w", op, err)
	}

	return &pb.RegRes{Token: token, SessionKey: ""}, nil
}
