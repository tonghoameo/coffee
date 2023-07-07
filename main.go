package main

import (
	"fmt"
	"os"

	"github.com/binbomb/storeCoffee/database"
	"github.com/binbomb/storeCoffee/middleware"
	"github.com/binbomb/storeCoffee/routes"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	fmt.Println("vim-go")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)
}
