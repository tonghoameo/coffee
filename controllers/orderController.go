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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

func GetOrders() gin.HandlerFunc {
	// c.Query = recordPerPage=2&page=2
	return func(c *gin.Context) {
		ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		rs, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cant fetch order collection"})
		}

		var orders []bson.M
		if err = rs.All(ctx, &orders); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, orders)

	}
}
func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		orderId := c.Param("order_id")
		var order models.Order
		err := orderCollection.FindOne(context.TODO(), bson.M{"order_id": orderId}).Decode(&order)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order id no found"})
		}

		c.JSON(http.StatusOK, order)
	}
}
func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		var table models.Table
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validatorErr := validate.Struct(order)
		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}
		if order.Table_id != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			if err != nil {
				msg := fmt.Sprint("TAble not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}
		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()
		result, insertErr := orderCollection.InsertOne(ctx, order)
		if insertErr != nil {
			msg := fmt.Sprintf("Order was not created")
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order
		var updateObj primitive.D
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		orderId := c.Param("order_id")

		if order.Table_id != nil {
			// check table exists
			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "table not found"})
				return
			}

			updateObj = append(updateObj, bson.E{"table_id", table.Table_id})
		}
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", order.Updated_at})

		upsert := true
		filter := bson.M{"order_id": orderId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		rs, err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Order update failed"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, rs)
	}
}

func OrderItemAdd(order models.Order) string {

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()
	orderCollection.InsertOne(ctx, order)
	defer cancel()
	return order.Order_id

}
