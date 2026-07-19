package auth

import (
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/auth/controller"
	"github.com/Hieu3z03/chat-api-golang/modules/auth/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	authController := do.MustInvoke[controller.AuthController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	authRoutes := server.Group("/api/auth")
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/refresh", authController.RefreshToken)
		authRoutes.POST("/logout", middlewares.Authenticate(jwtService), authController.Logout)
	}
}
