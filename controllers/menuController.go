package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/binbomb/storeCoffee/database"
	"github.com/binbomb/storeCoffee/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		rs, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error orcued while fetcing data find menu_id"})
		}

		var menus []bson.M
		if err = rs.All(ctx, &menus); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, menus)

	}
}
func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		menuId := c.Param("menu_id")
		var menu models.Menu
		var err = menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error orcued while fetcing data find menu_id"})
		}
		c.JSON(http.StatusOK, menu)
	}
}
func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		if err := c.BindJSON(&menu); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validatorErr := validate.Struct(menu)
		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()
		result, insertErr := menuCollection.InsertOne(ctx, menu)
		if insertErr != nil {
			msg := fmt.Sprintf("Menu item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id": menuId}

		var updateObj primitive.D
		if menu.Start_date != nil && menu.End_date != nil {
			if !intimeSpan(*menu.Start_date, *menu.End_date, time.Now()) {
				msg := "kindly retype the time"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
			}
		}
		updateObj = append(updateObj, bson.E{"start_date", menu.Start_date})
		updateObj = append(updateObj, bson.E{"end_date", menu.End_date})

		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{"name", menu.Name})
		}
		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{"category", menu.Category})
		}
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", menu.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		rs, err := menuCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)
		if err != nil {
			msg := "Menu update failed"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, rs)
	}
}
func intimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}
