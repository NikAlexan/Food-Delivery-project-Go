package router

import (
	"github.com/gin-gonic/gin"

	"food-delivery/order-service/gateway/internal/middleware"
	"food-delivery/order-service/gateway/internal/proxy"
)

func New(orderProxy *proxy.OrderProxy, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	orders := r.Group("/api/orders", middleware.JWTAuth(jwtSecret))
	{
		orders.POST("", orderProxy.CreateOrder)
		orders.GET("", orderProxy.ListUserOrders)
		orders.GET("/history", orderProxy.GetOrderHistory)
		orders.POST("/total", orderProxy.CalculateTotal)

		orders.GET("/:id", orderProxy.GetOrder)
		orders.PATCH("/:id/status", orderProxy.UpdateOrderStatus)
		orders.POST("/:id/cancel", orderProxy.CancelOrder)
		orders.POST("/:id/pay", orderProxy.ProcessPayment)
	}

	return r
}
