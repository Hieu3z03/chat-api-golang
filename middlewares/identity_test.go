package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireIdentityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		userID     string
		roleID     string
		wantStatus int
	}{
		{
			name:       "valid identity",
			userID:     "11111111-1111-4111-8111-111111111111",
			roleID:     "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "role id is optional for development clients",
			userID:     "11111111-1111-4111-8111-111111111111",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing user id",
			roleID:     "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid role id",
			userID:     "11111111-1111-4111-8111-111111111111",
			roleID:     "admin",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequireIdentityHeaders())
			router.GET("/test", func(ctx *gin.Context) {
				if _, ok := GetRequestIdentity(ctx); !ok {
					t.Fatal("identity was not stored in context")
				}
				ctx.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/test", nil)
			if test.userID != "" {
				request.Header.Set("x-user-id", test.userID)
			}
			if test.roleID != "" {
				request.Header.Set("x-user-role", test.roleID)
			}

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("expected status %d, got %d", test.wantStatus, response.Code)
			}
		})
	}
}
