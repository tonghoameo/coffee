package routes

import (
	controller "github.com/binbomb/storeCoffee/controllers"
	"github.com/gin-gonic/gin"
)

func FoodRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.GET("/foods", controller.GetFoods())
	incomingRoutes.GET("/foods/:food_id", controller.GetFood())
	incomingRoutes.POST("/foods", controller.CreateFood())
	incomingRoutes.PATCH("/foods/:food_id", controller.UpdateFood())
}
