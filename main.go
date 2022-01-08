package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	gin.DisableConsoleColor()

	// router with logger and crash free middleware
	router := gin.Default()
	router.Run(":8080")
}