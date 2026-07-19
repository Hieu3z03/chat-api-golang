package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Hieu3z03/chat-api-golang/middlewares"
	channelDTO "github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageController interface {
	Create(ctx *gin.Context)
	List(ctx *gin.Context)
}

type messageController struct {
	messages service.MessageService
}

func NewMessageController(messages service.MessageService) MessageController {
	return &messageController{messages: messages}
}

func (controller *messageController) Create(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	channelID, ok := parseChannelID(ctx)
	if !ok {
		return
	}

	var request dto.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid message payload", err.Error(), nil))
		return
	}

	message, err := controller.messages.Create(ctx.Request.Context(), channelID, identity.UserID, request)
	if errors.Is(err, dto.ErrInvalidMessageContent) {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed to create message", err.Error(), nil))
		return
	}
	if errors.Is(err, channelDTO.ErrNotChannelMember) {
		ctx.JSON(http.StatusForbidden, utils.BuildResponseFailed("channel access denied", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to create message", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess("message created", message))
}

func (controller *messageController) List(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	channelID, ok := parseChannelID(ctx)
	if !ok {
		return
	}

	limit, before, ok := parseListOptions(ctx)
	if !ok {
		return
	}

	messages, err := controller.messages.List(
		ctx.Request.Context(),
		channelID,
		identity.UserID,
		limit,
		before,
	)
	if errors.Is(err, channelDTO.ErrNotChannelMember) {
		ctx.JSON(http.StatusForbidden, utils.BuildResponseFailed("channel access denied", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to list messages", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("messages retrieved", messages))
}

func parseChannelID(ctx *gin.Context) (uuid.UUID, bool) {
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid channel id", err.Error(), nil))
		return uuid.Nil, false
	}

	return channelID, true
}

func parseListOptions(ctx *gin.Context) (int64, *time.Time, bool) {
	limit := int64(50)
	if value := ctx.Query("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 1 || parsed > 100 {
			ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid limit", "limit must be between 1 and 100", nil))
			return 0, nil, false
		}
		limit = parsed
	}

	var before *time.Time
	if value := ctx.Query("before"); value != "" {
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid before cursor", "before must use RFC3339 format", nil))
			return 0, nil, false
		}
		before = &parsed
	}

	return limit, before, true
}
