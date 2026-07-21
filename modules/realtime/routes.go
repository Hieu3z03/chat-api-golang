package realtime

import (
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/realtime/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	realtimeController := do.MustInvoke[controller.RealtimeController](injector)

	realtimeRoutes := server.Group("/api/realtime", middlewares.RequireIdentityHeaders())
	{
		realtimeRoutes.GET("/connection-token", realtimeController.ConnectionToken)
		realtimeRoutes.GET("/subscription-token", realtimeController.SubscriptionToken)
	}
}
