package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"users/internal/db"
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
	log.Info("Connected to database")

	rdb, err := rdb.NewRDB(log)
	if err != nil {
		log.Fatal("Failed to create redis", zap.Error(err))
	}
	defer rdb.Close()
	log.Info("Connected to redis")

	u := usersserver{log: log, db: db, rdb: rdb}
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

	token, err := security.GenerateToken(name+email, pswd)
	if err != nil {
		return nil, fmt.Errorf("%s: generate token: %w", op, err)
	}

	sk, err := u.rdb.NewSession(name+email, pswd)
	if err != nil {
		return nil, fmt.Errorf("%s: create session: %w", op, err)
	}

	defer func() {
		if err != nil {
			if err := u.rdb.Delete(sk); err != nil {
				u.log.Error("delete session",
					zap.String("op", op),
					zap.Error(err))
			}
		}
	}()

	ui, err := u.db.LogUser(name + email)
	if err != nil {
		u.log.Error(op, zap.Error(err))
		return nil, fmt.Errorf("%s: log user: %w", op, err)
	}

	if err = security.Check(pswd, ui.Pswd); err != nil {
		return nil, fmt.Errorf("%s: check password: %w", op, err)
	}

	return &pb.LogRes{Token: token, SessionKey: sk, UserId: ui.ID}, nil
}
