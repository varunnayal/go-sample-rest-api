package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var recipes []Recipe

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

func findRecipeIndex(id string) int {
	for index := 0; index < len(recipes); index++ {
		if recipes[index].ID == id {
			return index
		}
	}
	return -1
}

func _NewRecipeHandler(c *gin.Context) {
	var recipe Recipe;
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H {
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func _ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func _UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	index := findRecipeIndex(id)

	if index == -1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Recipe Not Found",
		})
		return
	}

	recipe.ID = recipes[index].ID
	recipe.PublishedAt = recipes[index].PublishedAt
	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

func _DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := findRecipeIndex(id)

	if index == -1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Recipe Not Found",
		})
		return
	}

	recipe := recipes[index]
	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, recipe)
}

func main() {
	gin.DisableConsoleColor()
	// initApp()

	// router with logger and crash free middleware
	router := gin.Default()
	router.GET("/recipes", _ListRecipesHandler)
	router.POST("/recipes", _NewRecipeHandler)
	router.PUT("/recipe/:id", _UpdateRecipeHandler)
	router.DELETE("/recipe/:id", _DeleteRecipeHandler)
	router.Run(":8080")
}

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)
}

