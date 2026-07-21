package controller

import (
	"errors"
	"net/http"

	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/realtime/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/utils"
	"github.com/gin-gonic/gin"
)

type RealtimeController interface {
	ConnectionToken(ctx *gin.Context)
	SubscriptionToken(ctx *gin.Context)
}

type realtimeController struct {
	realtime service.RealtimeService
}

type tokenResponse struct {
	Token string `json:"token"`
}

func NewRealtimeController(realtime service.RealtimeService) RealtimeController {
	return &realtimeController{realtime: realtime}
}

func (controller *realtimeController) ConnectionToken(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	token, err := controller.realtime.ConnectionToken(identity.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to issue connection token", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("connection token issued", tokenResponse{Token: token}))
}

func (controller *realtimeController) SubscriptionToken(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	channel := ctx.Query("channel")
	if channel == "" {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid realtime channel", "channel query parameter is required", nil))
		return
	}

	token, err := controller.realtime.SubscriptionToken(identity.UserID, channel)
	if errors.Is(err, service.ErrChannelAccessDenied) {
		ctx.JSON(http.StatusForbidden, utils.BuildResponseFailed("realtime channel access denied", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to issue subscription token", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("subscription token issued", tokenResponse{Token: token}))
}
