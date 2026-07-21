package controller

import (
	"errors"
	"net/http"

	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChannelController interface {
	Create(ctx *gin.Context)
	Get(ctx *gin.Context)
	List(ctx *gin.Context)
}

type channelController struct {
	channels service.ChannelService
}

func NewChannelController(channels service.ChannelService) ChannelController {
	return &channelController{channels: channels}
}

func (controller *channelController) Create(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)

	var request dto.CreateChannelRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid channel payload", err.Error(), nil))
		return
	}

	channel, err := controller.channels.Create(ctx.Request.Context(), identity.UserID, request)
	if errors.Is(err, dto.ErrInvalidChannelName) || errors.Is(err, dto.ErrMembersNotFound) {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed to create channel", err.Error(), nil))
		return
	}
	if errors.Is(err, dto.ErrChannelAlreadyExists) {
		ctx.JSON(http.StatusConflict, utils.BuildResponseFailed("failed to create channel", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to create channel", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess("channel created", channel))
}

func (controller *channelController) Get(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid channel id", err.Error(), nil))
		return
	}

	channel, err := controller.channels.Get(ctx.Request.Context(), identity.UserID, channelID)
	if errors.Is(err, dto.ErrNotChannelMember) {
		ctx.JSON(http.StatusForbidden, utils.BuildResponseFailed("channel access denied", err.Error(), nil))
		return
	}
	if errors.Is(err, dto.ErrChannelNotFound) {
		ctx.JSON(http.StatusNotFound, utils.BuildResponseFailed("channel not found", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to get channel", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("channel retrieved", channel))
}

func (controller *channelController) List(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	channels, err := controller.channels.List(ctx.Request.Context(), identity.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to list channels", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("channels retrieved", channels))
}
