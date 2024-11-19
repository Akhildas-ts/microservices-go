package admin

import (
	"context"
	"net/http"
	"strings"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin/pb"
	"github.com/gin-gonic/gin"
)

type AdminAuthMiddlewareConfig struct {
	svc *ServiceClient
}

// InitAdminAuthMiddleware initializes the middleware with the gRPC client.
func InitAdminAuthMiddleware(svc *ServiceClient) AdminAuthMiddlewareConfig {
	return AdminAuthMiddlewareConfig{svc}
}

// AuthRequired checks if the admin token is valid.
func (c *AdminAuthMiddlewareConfig) AuthRequired(ctx *gin.Context) {
	// Retrieve the authorization header.
	authorization := ctx.Request.Header.Get("authorization")
	if authorization == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	// Extract the Bearer token.
	token := strings.Split(authorization, "Bearer ")
	if len(token) < 2 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token missing"})
		return
	}

	// Validate the token using the gRPC service.
	res, err := c.svc.Client.Validate(context.Background(), &pb.ValidateRequest{
		Token: token[1],
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed"})
		return
	}

	// Set the user ID (admin ID) in the context for downstream use.
	ctx.Set("adminId", res.AdminId)

	// Allow the request to proceed.
	ctx.Next()
}
