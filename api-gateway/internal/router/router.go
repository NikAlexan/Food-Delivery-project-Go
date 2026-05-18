package router

import (
	"github.com/gin-gonic/gin"

	"food-delivery/api-gateway/internal/middleware"
	"food-delivery/api-gateway/internal/proxy"
)

func New(userProxy *proxy.UserProxy, deliveryProxy *proxy.DeliveryProxy, restaurantProxy *proxy.RestaurantProxy, orderProxy *proxy.OrderProxy, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimit())
	r.Use(middleware.RequestLogger())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

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
			auth.GET("/my", restaurantProxy.GetMyRestaurant)
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
		delivery.POST("/register", deliveryProxy.RegisterDriver)
		delivery.GET("/my-driver", deliveryProxy.GetMyDriver)
		delivery.POST("/assign", deliveryProxy.AssignDriver)
		delivery.PATCH("/location", deliveryProxy.UpdateDriverLocation)
		delivery.POST("/:id/complete", deliveryProxy.CompleteDelivery)
		delivery.GET("/drivers/:driverId", deliveryProxy.ListDriverDeliveries)
		delivery.GET("/history", deliveryProxy.GetDeliveryHistory)
		delivery.GET("/:id", deliveryProxy.GetDelivery)
		delivery.GET("/:id/track", deliveryProxy.TrackDelivery)
		delivery.POST("/:id/cancel", deliveryProxy.CancelDelivery)
	}

	// ── Orders (auth required) ────────────────────────────────────────────────
	orders := r.Group("/api/orders", middleware.JWTAuth(jwtSecret))
	{
		orders.POST("", orderProxy.CreateOrder)
		orders.GET("", orderProxy.ListUserOrders)
		orders.GET("/history", orderProxy.GetOrderHistory)
		orders.POST("/calculate", orderProxy.CalculateTotal)
		orders.GET("/:id", orderProxy.GetOrder)
		orders.PATCH("/:id/status", orderProxy.UpdateOrderStatus)
		orders.POST("/:id/cancel", orderProxy.CancelOrder)
		orders.POST("/:id/payment", orderProxy.ProcessPayment)
	}

	return r
}
