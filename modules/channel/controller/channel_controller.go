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
	Delete(ctx *gin.Context)
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
		middlewares.RecordError(ctx, err)
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid channel payload", err.Error(), nil))
		return
	}

	channel, err := controller.channels.Create(ctx.Request.Context(), identity.UserID, request)
	middlewares.RecordError(ctx, err)
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
		middlewares.RecordError(ctx, err)
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid channel id", err.Error(), nil))
		return
	}

	channel, err := controller.channels.Get(ctx.Request.Context(), identity.UserID, channelID)
	middlewares.RecordError(ctx, err)
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
		middlewares.RecordError(ctx, err)
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to list channels", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("channels retrieved", channels))
}

func (controller *channelController) Delete(ctx *gin.Context) {
	// 1. Trích xuất thông tin định danh
	identity, exists := middlewares.GetRequestIdentity(ctx)
	if !exists {
		// Gọi đúng hàm BuildResponseFailed với đầy đủ tham số (Message, Error, Data)
		response := utils.BuildResponseFailed("Yêu cầu xác thực tài khoản", "Unauthorized", nil)
		ctx.JSON(http.StatusUnauthorized, response)
		return
	}

	// 2. Lấy và kiểm tra định dạng Channel ID từ URL
	idParam := ctx.Param("channel_id")
	channelID, err := uuid.Parse(idParam)
	if err != nil {
		response := utils.BuildResponseFailed("Định dạng Channel ID không hợp lệ", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	// 3. Gọi tầng Service xử lý (Đổi tên biến context của Go thành goCtx để tránh trùng với ctx của Gin)
	goCtx := ctx.Request.Context()
	err = controller.channels.DeleteChannel(goCtx, channelID, identity.UserID)
	if err != nil {
		// Xử lý lỗi không tìm thấy channel
		if err.Error() == "channel không tồn tại" {
			response := utils.BuildResponseFailed("Không thể xóa channel", err.Error(), nil)
			ctx.JSON(http.StatusNotFound, response)
			return
		}

		// Xử lý lỗi phân quyền
		if err.Error() == "bạn không có quyền xóa channel này" {
			response := utils.BuildResponseFailed("Từ chối truy cập", err.Error(), nil)
			ctx.JSON(http.StatusForbidden, response)
			return
		}

		// Xử lý các lỗi hệ thống khác
		response := utils.BuildResponseFailed("Xóa channel thất bại", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	// 4. Trả về phản hồi thành công thống nhất cấu trúc
	successResponse := utils.BuildResponseSuccess("Xóa channel thành công", nil)
	ctx.JSON(http.StatusOK, successResponse)
}
