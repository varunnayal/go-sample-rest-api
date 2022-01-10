package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"recipes-api/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RecipesHandler struct
type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

// NewRecipeHandler create new structure
func NewRecipeHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// GetRecipeByID handler
func (handler *RecipesHandler) GetRecipeByID(c *gin.Context) {
	objectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id",
		})
		return
	}

	var recipe models.Recipe
	err = handler.collection.FindOne(handler.ctx, bson.M{"_id": objectID}).Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to find data",
			"reason":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

func (handler *RecipesHandler) clearRecipesListCache() {
	log.Println("clearing cache")
	if err := handler.redisClient.Del("002-recipe-app:recipes").Err(); err != nil {
		log.Println("[ERROR] cache clear ", err.Error())
	}
}

// NewRecipeHandler handler
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	if _, err := handler.collection.InsertOne(handler.ctx, recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Error adding new recipe",
			"reason": err.Error(),
		})
		return
	}

	handler.clearRecipesListCache()
	c.JSON(http.StatusOK, recipe)
}

// ListRecipesHandler handler
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	cacheResult, err := handler.redisClient.Get("002-recipe-app:recipes").Result()
	if err == nil {
		log.Println("cache_hit recipes")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(cacheResult), &recipes)
		c.JSON(http.StatusOK, recipes)
		return
	} else if err != redis.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// bring it from MongoDB and save it in cache
	log.Println("cache_miss recipes")

	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer cur.Close(handler.ctx)

	recipes := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	data, _ := json.Marshal(recipes)
	if _, err := handler.redisClient.Set("002-recipe-app:recipes", string(data), 2 * time.Minute).Result(); err != nil {
		log.Println("[ERROR] Redis::Set", err)
	}

	c.JSON(http.StatusOK, recipes)
}

// UpdateRecipeHandler handler
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid id",
			"reason": err.Error(),
		})
		return
	}
	updateRes, err := handler.collection.UpdateOne(handler.ctx, bson.M{"_id": objectID}, bson.D{
		{"$set", bson.D{
			{"name", recipe.Name},
			{"tags", recipe.Tags},
			{"ingredients", recipe.Ingredients},
			{"instructions", recipe.Instructions},
		}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in updating", "reason": err.Error()})
		return
	}

	log.Println("Update Result: ", updateRes)
	handler.clearRecipesListCache()
	c.JSON(http.StatusOK, updateRes)
}

// DeleteRecipeHandler handler
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	recipeID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid id",
		})
		return
	}

	delResult, err := handler.collection.DeleteOne(handler.ctx, bson.M{"_id": recipeID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if delResult.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Invalid id",
		})
		return
	}
	handler.clearRecipesListCache()
	c.JSON(http.StatusOK, delResult)
}

// SearchRecipeHandler handler
func (handler *RecipesHandler) SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")

	cur, err := handler.collection.Find(handler.ctx, bson.M{"tags": bson.M{"$in": []string{tag}}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Error adding new recipe",
			"reason": err.Error(),
		})
		return
	}
	defer cur.Close(handler.ctx)
	filteredRecipes := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		cur.Decode(&recipe)
		filteredRecipes = append(filteredRecipes, recipe)
	}
	c.JSON(http.StatusOK, filteredRecipes)
}
