package main

import (
	"github.com/gin-gonic/gin"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

func main() {
	gin.DisableConsoleColor()

	// router with logger and crash free middleware
	router := gin.Default()
	router.Run(":8080")
}