package main

import (
	"gourses/internal/routers"
)

func main() {
	srv := routers.Init()
	srv.Run(":8080")
}
