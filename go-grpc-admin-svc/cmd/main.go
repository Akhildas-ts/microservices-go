package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/config"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/db"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/pb"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/services"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/utils"
	"google.golang.org/grpc"
)

func main() {

	fmt.Println("inside the admin service ... ")
	// Load configuration
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("Failed to load config:", err)
	}

	// Initialize database
	h := db.Init(c.DBUrl)

	// Initialize JWT wrapper
	jwt := utils.JwtWrapper{
		SecretKey:       c.JWTSecretKey,
		Issuer:          "go-grpc-auth-svc",
		ExpirationHours: 24 * 365, // 1 year
	}

	// Listen on the specified port
	lis, err := net.Listen("tcp", c.Port)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	fmt.Println("Auth Service running on", c.Port)

	// Initialize the server with handlers and JWT
	s := services.Server{
		H:   h,
		Jwt: jwt,
	}

	// Create new gRPC server instance
	grpcServer := grpc.NewServer()

	// Register the AuthService server
	pb.RegisterAdminServiceServer(grpcServer, &s)

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
