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

type OrderItemPack struct {
	Table_id    *string
	Order_Items []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		rs, err := orderItemCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cant fetch orderItem collection"})
		}

		var orderItems []bson.M
		if err = rs.All(ctx, &orderItems); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, orderItems)
	}
}
func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	// orderItem and relationship food one and many
	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}
	// {"as", "food"} get data from food merge in order item rename as like food
	lookupStage := bson.D{{"$lookup", bson.D{{"from", "food"}, {"localField", "food_id"}, {"foreignField", "food_id"}, {"as", "food"}}}}
	unwindStage := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullAndEmptyArrays", true}}}}

	// order and orderitem
	lookupOrderStage := bson.D{{"$lookup", bson.D{{"food", "order"}, {"localField", "order_id"}, {"foreignField", "order_id"}, {"as", "order"}}}}
	unwindOrderStage := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullAndEmptyArrays", "true"}}}}

	// order and table
	lookupTableStage := bson.D{{"$lookup", bson.D{{"from", "table"}, {"localField", "order.table_id"}, {"foreignField", "table_id"}, {"as", "table"}}}}
	// view table $table
	unwindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullAndEmptyArrays", "true"}}}}

	// projectStage
	projectStage := bson.D{
		{
			"$projectStage", bson.D{
				{"id", 0},
				{"amount", "$food.price"},
				{"total_count", 1},
				{"food_name", "$food.name"},
				{"food_image", "$food.food_image"},
				{"table_number", "$table.table_number"},
				{"table_id", "$table.table_id"},
				{"order_id", "$order.order_id"},
				{"price", "$food.price"},
				{"quantity", 1},
			}}}
	groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"order_id", "$order_id"}, {"table_id", "$table_id"}, {"table_number", "$table_number"}}}, {"payment_due", bson.D{{"$sum", "$amount"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"order_items", bson.D{{"$push", "$$ROOT"}}}}}}

	projectStage2 := bson.D{
		{"$project", bson.D{
			{"id", 0},
			{"payment_due", 1},
			{"total_count", 1},
			{"table_number", "$_id.table_number"},
			{"order_items", 1},
		}}}
	rs, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if err = rs.All(ctx, &OrderItems); err != nil {
		panic(err)
	}
	defer cancel()
	return OrderItems, nil

}
func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		orderItemId := c.Param("order_item_id")
		var orderItem models.OrderItem
		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order item id no found"})
		}

		c.JSON(http.StatusOK, orderItem)

	}
}
func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		var orderItemPack OrderItemPack
		fmt.Println("chack bindjson orderitempack")
		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if orderItemPack.Table_id == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "bind json wron table_id nil"})
			return

		}
		order.Order_date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		orderItemsToBeInserted := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_id := OrderItemAdd(order)

		for _, orderItem := range orderItemPack.Order_Items {
			orderItem.Order_id = order_id
			validatorErr := validate.Struct(orderItem)
			if validatorErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
				return
			}
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.ID = primitive.NewObjectID()
			orderItem.Order_id = order.ID.Hex()
			var num = toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		insertOrderItems, insertErr := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
		if insertErr != nil {
			msg := fmt.Sprintf("Order was not created")
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, insertOrderItems)

	}
}
func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var orderItem models.OrderItem

		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		orderItemId := c.Param("order_item_id")
		filter := bson.M{"order_item_id": orderItemId}

		var updateObj primitive.D
		if orderItem.Unit_price != nil {
			updateObj = append(updateObj, bson.E{"unit_price", *&orderItem.Unit_price})
		}
		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", *orderItem.Quantity})
		}
		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{"food_id", *orderItem.Food_id})
		}
		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"Updated_at", orderItem.Updated_at})
		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		rs, err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{{"$set", updateObj}},
			&opt,
		)

		if err != nil {
			msg := "order item failed"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		defer cancel()
		c.JSON(http.StatusOK, rs)
	}
}
