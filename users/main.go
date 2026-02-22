package main

import (
	"users/internal/db"

	pb "github.com/Votline/Gourses/protos/generated-users"
	"go.uber.org/zap"
)

type usersserver struct {
	log *zap.Logger
	db  *db.DB
	pb.UnimplementedUsersServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	defer log.Sync()

	db, err := db.NewDB(log)
	if err != nil {
		log.Fatal("Failed to create database", zap.Error(err))
	}
	defer db.Close()

	// u := user{log: log, db: db}
}
