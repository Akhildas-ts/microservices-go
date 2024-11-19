package admin

import (
	"fmt"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin/routes"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, c *config.Config) *ServiceClient {

	svc := &ServiceClient{Client: InitServiceClient(c)}

	routes := r.Group("/admin")
	routes.POST("/login", svc.LoginAdmin)
	routes.POST("/signup", svc.SignupAdmin)

	return svc
}

func (svc *ServiceClient) LoginAdmin(ctx *gin.Context) {
	routes.LoginAdmin(ctx, svc.Client)
}

func (svc *ServiceClient) SignupAdmin(ctx *gin.Context) {

	fmt.Println("admin signup called ..  from routes.go")
	routes.SingupAdmin(ctx, svc.Client)
}
