package cart

import (
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/cart/routes"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/config"
	"github.com/gin-gonic/gin"
)

func CartRoutes(r *gin.Engine, c *config.Config) *ServiceClient {
	svc := &ServiceClient{Client: InitServiceClient(c)}

	routes := r.Group("/cart")
	routes.POST("/addcart", svc.AddCart)
	routes.GET("/getcart", svc.GetCart)

	return svc

}

func (svc *ServiceClient) AddCart(ctx *gin.Context) {

	routes.AddToCart(ctx, svc.Client)

}

func (svc *ServiceClient) GetCart(ctx *gin.Context) {
	routes.GetCart(ctx, svc.Client)
}
