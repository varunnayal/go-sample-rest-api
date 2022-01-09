package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	handlers "recipes-api/handlers"
	"recipes-api/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var client *mongo.Client
var collection *mongo.Collection
var recipesHandler *handlers.RecipesHandler

func main() {
	gin.DisableConsoleColor()
	// initApp()

	// router with logger and crash free middleware
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		var r *gin.RouterGroup
		r = v1.Group("/recipe")
		{
			r.GET("/:id", recipesHandler.GetRecipeByID)
			r.PUT("/:id", recipesHandler.UpdateRecipeHandler)
			r.DELETE("/:id", recipesHandler.DeleteRecipeHandler)
		}

		r = v1.Group("/recipes")
		{
			r.POST("/", recipesHandler.NewRecipeHandler)
			r.GET("/", recipesHandler.ListRecipesHandler)
			r.GET("/search", recipesHandler.SearchRecipeHandler)
		}
	}

	router.Run(":8080")
}

func _loadFromJSON(recipes []models.Recipe) {
	recipes = make([]models.Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)

	var recipeList []interface{}
	for _, recipe := range recipes {
		recipe.ID = primitive.NewObjectID()
		recipeList = append(recipeList, recipe)
	}

	insertManyResult, err := collection.InsertMany(context.TODO(), recipeList)
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
	var client *mongo.Client
	var err error

	for true {
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
		if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
			log.Println("<<<<<<<<<< ERROR >>>>>>>>>>>>")
			log.Println(err)
			time.Sleep(5 * time.Duration(time.Second))
		} else {
			log.Println("Connected to MongoDB")
			break
		}
	}
	return client
}

func init() {
	client = _connectMongo()
	collection = client.Database(os.Getenv("MONGO_DB")).Collection("recipes")

	recipesHandler = handlers.NewRecipeHandler(ctx, collection)
	// _loadFromJSON(recipes)
	_countDocuments()
}
