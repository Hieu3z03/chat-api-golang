package channel

import (
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	channelController := do.MustInvoke[controller.ChannelController](injector)

	channelRoutes := server.Group("/api/channels", middlewares.RequireIdentityHeaders())
	{
		channelRoutes.POST("", channelController.Create)
		channelRoutes.GET("", channelController.List)
		channelRoutes.GET("/:channel_id", channelController.Get)
		channelRoutes.DELETE("/:channel_id", channelController.Delete)
	}
}
