package product

import (
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/config"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/product/routes"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, c *config.Config, adminSvc *admin.ServiceClient) {
	a := admin.InitAdminAuthMiddleware(adminSvc)

	svc := &ServiceClient{
		Client: InitServiceClient(c),
	}

	routes := r.Group("/product")
	routes.Use(a.AuthRequired)
	routes.POST("/", svc.CreateProduct)
	routes.GET("/:id", svc.FindOne)
}

func (svc *ServiceClient) FindOne(ctx *gin.Context) {
	routes.FineOne(ctx, svc.Client)
}

func (svc *ServiceClient) CreateProduct(ctx *gin.Context) {
	routes.CreateProduct(ctx, svc.Client)
}
