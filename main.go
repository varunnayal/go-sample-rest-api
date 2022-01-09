package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var err error
var client *mongo.Client
var collection *mongo.Collection
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

func _SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")

	filteredRecipes := make([]Recipe, 0)

	for i := 0; i < len(recipes); i++ {
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				filteredRecipes = append(filteredRecipes, recipes[i])
				break;
			}
		}
	}

	c.JSON(http.StatusOK, filteredRecipes)
}

func main() {
	gin.DisableConsoleColor()
	// initApp()

	// router with logger and crash free middleware
	router := gin.Default()
	router.GET("/recipes", _ListRecipesHandler)
	router.GET("/recipes/search", _SearchRecipeHandler)
	router.POST("/recipes", _NewRecipeHandler)
	router.PUT("/recipe/:id", _UpdateRecipeHandler)
	router.DELETE("/recipe/:id", _DeleteRecipeHandler)
	router.Run(":8080")
}

func _loadFromJSON(recipes []Recipe) {
	var recipeList []interface{}
	for _, 	recipe := range recipes {
		recipeList = append(recipeList, recipe)
	}


	insertManyResult, err := collection.InsertMany(context.TODO(), recipeList);
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Added recipes: ", len(insertManyResult.InsertedIDs))
}

func _countDocuments() {
	n, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Available recipes: ", n)

}

func _connectMongo() *mongo.Client {
	ctx = context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	for true {
		if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
			log.Println("<<<<<<<<<< ERROR >>>>>>>>>>>>")
			log.Println(err)
			time.Sleep(5 * time.Duration(time.Second))
			// log.Fatal(err)
		} else {
			log.Println("Connected to MongoDB")
			break
		}
	}
	return client
}

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)

	client = _connectMongo()
	collection = client.Database(os.Getenv("MONGO_DB")).Collection("recipes")


	// _loadFromJSON(recipes)
	_countDocuments()
}

