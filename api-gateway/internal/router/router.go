package router

import (
	"github.com/gin-gonic/gin"

	"food-delivery/api-gateway/internal/middleware"
	"food-delivery/api-gateway/internal/proxy"
)

func New(userProxy *proxy.UserProxy, deliveryProxy *proxy.DeliveryProxy, restaurantProxy *proxy.RestaurantProxy, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	// ── Users ─────────────────────────────────────────────────────────────────
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

	// ── Restaurants (public read, authenticated write) ─────────────────────────
	restaurants := r.Group("/api/restaurants")
	{
		restaurants.GET("", restaurantProxy.ListRestaurants)
		restaurants.GET("/search", restaurantProxy.SearchRestaurants)
		restaurants.GET("/:id", restaurantProxy.GetRestaurant)
		restaurants.GET("/:id/menu", restaurantProxy.GetMenu)

		auth := restaurants.Group("", middleware.JWTAuth(jwtSecret))
		{
			auth.POST("", restaurantProxy.CreateRestaurant)
			auth.PUT("/:id", restaurantProxy.UpdateRestaurant)
			auth.DELETE("/:id", restaurantProxy.DeleteRestaurant)
			auth.POST("/:id/menu", restaurantProxy.CreateMenuItem)
			auth.PUT("/:id/menu/:item_id", restaurantProxy.UpdateMenuItem)
			auth.DELETE("/:id/menu/:item_id", restaurantProxy.DeleteMenuItem)
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
