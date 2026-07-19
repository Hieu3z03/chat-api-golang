package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/user/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/user/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController interface {
	Sync(ctx *gin.Context)
	Me(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	Search(ctx *gin.Context)
}

type userController struct {
	users service.UserService
}

func NewUserController(users service.UserService) UserController {
	return &userController{users: users}
}

func (controller *userController) Sync(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)

	var request dto.SyncUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid user payload", err.Error(), nil))
		return
	}

	user, err := controller.users.Sync(ctx.Request.Context(), identity.UserID, request)
	if errors.Is(err, dto.ErrUsernameTaken) {
		ctx.JSON(http.StatusConflict, utils.BuildResponseFailed("failed to sync user", err.Error(), nil))
		return
	}
	if errors.Is(err, dto.ErrInvalidUserProfile) {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed to sync user", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to sync user", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("user synchronized", user))
}

func (controller *userController) Me(ctx *gin.Context) {
	identity, _ := middlewares.GetRequestIdentity(ctx)
	controller.respondWithUser(ctx, identity.UserID)
}

func (controller *userController) GetByID(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid user id", err.Error(), nil))
		return
	}

	controller.respondWithUser(ctx, userID)
}

func (controller *userController) Search(ctx *gin.Context) {
	limit := 20
	if value := ctx.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 50 {
			ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("invalid limit", "limit must be between 1 and 50", nil))
			return
		}
		limit = parsed
	}

	users, err := controller.users.Search(ctx.Request.Context(), ctx.Query("search"), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to list users", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("users retrieved", users))
}

func (controller *userController) respondWithUser(ctx *gin.Context, userID uuid.UUID) {
	user, err := controller.users.GetByID(ctx.Request.Context(), userID)
	if errors.Is(err, dto.ErrUserNotFound) {
		ctx.JSON(http.StatusNotFound, utils.BuildResponseFailed("user not found", err.Error(), nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed to get user", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("user retrieved", user))
}
