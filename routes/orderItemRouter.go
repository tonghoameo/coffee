package routes

import (
	controller "github.com/binbomb/storeCoffee/controllers"
	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.GET("/orderitems", controller.GetOrderItems())
	incomingRoutes.GET("/orderitems/:orderitem_id", controller.GetOrderItem())
	incomingRoutes.GET("/itemsfromorder/:order_id", controller.GetOrderItemsByOrder())
	incomingRoutes.POST("/orderitems", controller.CreateOrderItem())
	incomingRoutes.PATCH("/orderitems/:orderitem_id", controller.UpdateOrderItem())
}
