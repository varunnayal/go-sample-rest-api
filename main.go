package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	handlers "recipes-api/handlers"
	"recipes-api/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var client *mongo.Client
var collection *mongo.Collection
var redisClient *redis.Client
var recipesHandler *handlers.RecipesHandler

func main() {
	gin.DisableConsoleColor()
	// initApp()

	// router with logger and crash free middleware
	// router := gin.Default()
	router := gin.New()
	router.Use(gin.Logger())

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
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")).SetMaxPoolSize(100).SetAppName("gin-recipe-app"))
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

func _connectRedis() *redis.Client {
	redisOpts := &redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PWD"),
		DB:       0,
	}
	redisClient := redis.NewClient(redisOpts)

	log.Println("redisClient = ", redisClient)

	pingRes, err := redisClient.Ping().Result()
	if err != nil {
		log.Println("Redis Error: ", err)
	} else {
		fmt.Println("Redis ping -> ", pingRes)
	}
	return redisClient
}

func init() {
	// TODO parallel connection mongo and redis
	client = _connectMongo()
	redisClient = _connectRedis()
	collection = client.Database(os.Getenv("MONGO_DB")).Collection("recipes")

	recipesHandler = handlers.NewRecipeHandler(ctx, collection, redisClient)
	// _loadFromJSON(recipes)
	_countDocuments()
}
