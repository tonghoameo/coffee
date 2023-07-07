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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		rs, err := tableCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cant fetch table collection"})
		}

		var tables []bson.M
		if err = rs.All(ctx, &tables); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, tables)

	}
}
func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		tableId := c.Param("table_id")
		var table models.Table
		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "table item id no found"})
		}

		c.JSON(http.StatusOK, table)

	}
}
func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var table models.Table
		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validatorErr := validate.Struct(table)
		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()
		rs, insertErr := tableCollection.InsertOne(ctx, table)
		if insertErr != nil {
			msg := fmt.Sprintf("Table was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, rs)

	}
}
func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var table models.Table
		tableId := c.Param("table_id")
		if err := c.BindJSON(&table); err != nil {
			defer cancel()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var updateObj primitive.D
		if table.Table_of_guests != nil {
			updateObj = append(updateObj, bson.E{"table_of_guests", table.Table_of_guests})
		}
		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{"table_num", table.Table_number})
		}
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		filter := bson.M{"table_id": tableId}
		rs, err := tableCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "update table is failed"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		defer cancel()
		c.JSON(http.StatusOK, rs)
	}
}
