package user

import (
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/auth/service"
	"github.com/Hieu3z03/chat-api-golang/modules/user/controller"
	"github.com/Hieu3z03/chat-api-golang/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	userController := do.MustInvoke[controller.UserController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	userRoutes := server.Group("/api/user")
	{
		userRoutes.GET("", userController.GetAllUser)
		userRoutes.GET("/me", middlewares.Authenticate(jwtService), userController.Me)
		userRoutes.PUT("/:id", middlewares.Authenticate(jwtService), userController.Update)
		userRoutes.DELETE("/:id", middlewares.Authenticate(jwtService), userController.Delete)
	}
}
