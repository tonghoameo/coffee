package controller

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/binbomb/storeCoffee/database"
	"github.com/binbomb/storeCoffee/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Query = recordPerPage=2&page=2
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIdx := (page - 1) * recordPerPage
		startIdx, err = strconv.Atoi(c.Query("startIdx"))
		matchStage := bson.D{{"$match", bson.D{}}}
		groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_Count", 1},
					{"food_items", bson.D{{"$slice", []interface{}{"$data", startIdx, recordPerPage}}}},
				}}}

		rs, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error when aggrate foods"})
		}
		var allFoods []bson.M
		if err = rs.All(ctx, &allFoods); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allFoods[0])

	}
}
func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food
		var err = foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error orcued while fetcing data find food_id"})
		}
		c.JSON(http.StatusOK, food)
	}
}
func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food
		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validatorErr := validate.Struct(food)
		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()

		if err != nil {
			msg := fmt.Sprintf("menu was found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num
		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil {
			msg := fmt.Sprintf("Food item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food
		defer cancel()
		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{"name", food.Name})
		}
		if food.Price != nil {
			updateObj = append(updateObj, bson.E{"price", food.Price})

		}
		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{"food_image", food.Food_image})

		}
		if food.Menu_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			if err != nil {
				msg := fmt.Sprint(" food can not in menu")
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"menu_id", food.Menu_id})

		}
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", food.Updated_at})
		upsert := true
		filter := bson.M{"food_id": foodId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		rs, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprint("food item cannot update")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, rs)
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
