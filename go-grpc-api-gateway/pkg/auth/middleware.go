package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/auth/pb"
	"github.com/gin-gonic/gin"
)

type AuthMiddlewareConfig struct {
	svc *ServiceClient
}

func InitAuthMiddleware(svc *ServiceClient) AuthMiddlewareConfig {
	return AuthMiddlewareConfig{svc}
}
func (c *AuthMiddlewareConfig) AuthRequired(ctx *gin.Context) {
	authorization := ctx.Request.Header.Get("authorization")

	if authorization == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header missing",
		})
		ctx.Abort()
		return
	}

	token := strings.Split(authorization, "Bearer ")

	if len(token) < 2 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid authorization token format",
		})
		ctx.Abort()
		return
	}

	res, err := c.svc.Client.Validate(context.Background(), &pb.ValidateRequest{
		Token: token[1],
	})

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token validation failed",
			"details": err.Error(),
		})
		ctx.Abort()
		return
	}

	if res.Status != http.StatusOK {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		ctx.Abort()
		return
	}

	ctx.Set("userId", res.UserId)
	ctx.Next()
}
