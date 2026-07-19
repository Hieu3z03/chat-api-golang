package user

import (
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/user/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	userController := do.MustInvoke[controller.UserController](injector)

	userRoutes := server.Group("/api/users", middlewares.RequireIdentityHeaders())
	{
		userRoutes.PUT("/me", userController.Sync)
		userRoutes.GET("/me", userController.Me)
		userRoutes.GET("", userController.Search)
		userRoutes.GET("/:user_id", userController.GetByID)
	}
}
