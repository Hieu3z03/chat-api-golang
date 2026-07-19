package message

import (
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/message/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	messageController := do.MustInvoke[controller.MessageController](injector)

	messageRoutes := server.Group("/api/channels/:channel_id/messages", middlewares.RequireIdentityHeaders())
	{
		messageRoutes.POST("", messageController.Create)
		messageRoutes.GET("", messageController.List)
	}
}
