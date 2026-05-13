package router

import (
	"github.com/gin-gonic/gin"

	"food-delivery/api-gateway/internal/middleware"
	"food-delivery/api-gateway/internal/proxy"
)

func New(userProxy *proxy.UserProxy, deliveryProxy *proxy.DeliveryProxy, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	users := r.Group("/api/users")
	{
		users.POST("/register", userProxy.Register)
		users.POST("/login", userProxy.Login)
		users.POST("/refresh", userProxy.RefreshToken)

		auth := users.Group("", middleware.JWTAuth(jwtSecret))
		{
			auth.GET("/profile", userProxy.GetProfile)
			auth.PUT("/profile", userProxy.UpdateProfile)
			auth.DELETE("/me", userProxy.DeleteUser)
			auth.POST("/addresses", userProxy.AddAddress)
			auth.GET("/addresses", userProxy.GetAddresses)
		}
	}

	delivery := r.Group("/api/delivery", middleware.JWTAuth(jwtSecret))
	{
		delivery.POST("/assign", deliveryProxy.AssignDriver)
		delivery.PATCH("/location", deliveryProxy.UpdateDriverLocation)
		delivery.POST("/:id/complete", deliveryProxy.CompleteDelivery)
		delivery.GET("/drivers/:driverId", deliveryProxy.ListDriverDeliveries)
		delivery.GET("/history", deliveryProxy.GetDeliveryHistory)
		delivery.GET("/:id", deliveryProxy.GetDelivery)
		delivery.GET("/:id/track", deliveryProxy.TrackDelivery)
		delivery.POST("/:id/cancel", deliveryProxy.CancelDelivery)
	}

	return r
}
