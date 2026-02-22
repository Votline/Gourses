package main

import (
	"os"

	"gateway/internal/routers"

	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewDevelopment()
	srv := routers.Init(log)
	srv.Run(":" + os.Getenv("API_PORT"))
}
