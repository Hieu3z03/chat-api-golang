package middlewares

import (
	"net/http"

	"github.com/Hieu3z03/chat-api-golang/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const identityContextKey = "request_identity"

type RequestIdentity struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

func RequireIdentityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, err := uuid.Parse(ctx.GetHeader("x-user-id"))
		if err != nil {
			response := utils.BuildResponseFailed(
				"invalid request identity",
				"x-user-id header must be a valid UUID",
				nil,
			)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		roleID := uuid.Nil
		if value := ctx.GetHeader("x-user-role"); value != "" {
			roleID, err = uuid.Parse(value)
			if err != nil {
				response := utils.BuildResponseFailed(
					"invalid request identity",
					"x-user-role header must be a valid UUID when provided",
					nil,
				)
				ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
				return
			}
		}

		ctx.Set(identityContextKey, RequestIdentity{
			UserID: userID,
			RoleID: roleID,
		})
		ctx.Next()
	}
}

func GetRequestIdentity(ctx *gin.Context) (RequestIdentity, bool) {
	value, exists := ctx.Get(identityContextKey)
	if !exists {
		return RequestIdentity{}, false
	}

	identity, ok := value.(RequestIdentity)
	return identity, ok
}
