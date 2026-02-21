package main

import (
	"gourses/internal/routers"
	"os"
)

func main() {
	srv := routers.Init()
	srv.Run(":"+os.Getenv("API_PORT"))
}
