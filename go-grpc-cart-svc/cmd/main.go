package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/client"
	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/config"
	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/db"
	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/pb"
	service "github.com/Akhildas-ts/go-grpc-cart-svc/pkg/services"
	"google.golang.org/grpc"
)

func main() {
	c, err := config.LoadConfig()

	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	h := db.Init(c.DBUrl)

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	productSvc := client.InitProductServiceClient(c.ProductSvcUrl)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	s := service.Server{
		H:          h,
		ProductSvc: productSvc,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterCartServer(grpcServer, &s)
	fmt.Println("server  runing ")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
