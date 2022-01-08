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

	index := -1
	for i:=0; i<len(recipes); i++ {
		if recipes[i].ID == id {
			index = i;
			break;
		}
	}

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

func main() {
	gin.DisableConsoleColor()
	// initApp()

	// router with logger and crash free middleware
	router := gin.Default()
	router.GET("/recipes", _ListRecipesHandler)
	router.POST("/recipes", _NewRecipeHandler)
	router.PUT("/recipe/:id", _UpdateRecipeHandler)
	router.Run(":8080")
}

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)
}

