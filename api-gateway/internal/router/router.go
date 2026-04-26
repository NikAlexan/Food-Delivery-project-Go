package router

import (
	"github.com/gin-gonic/gin"

	"food-delivery/api-gateway/internal/middleware"
	"food-delivery/api-gateway/internal/proxy"
)

func New(userProxy *proxy.UserProxy, jwtSecret string) *gin.Engine {
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

	return r
}