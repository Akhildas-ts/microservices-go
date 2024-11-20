package main

import (
	"log"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/auth"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/config"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/order"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/product"
	"github.com/gin-gonic/gin"
)

func main() {
	c, err := config.LoadConfig()

	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	r := gin.Default()

	authSvc := *auth.RegisterRoutes(r, &c)
	adminSvc := *admin.RegisterRoutes(r, &c)
	product.RegisterRoutes(r, &c, &adminSvc)
	order.RegisterRoutes(r, &c, &authSvc)

	r.Run(c.Port)
}
