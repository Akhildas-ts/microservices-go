package admin

import (
	"fmt"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin/pb"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/config"
	"google.golang.org/grpc"
)

type ServiceClient struct {
	Client pb.AdminServiceClient
}

func InitServiceClient(c *config.Config) pb.AdminServiceClient {
	// using WithInsecure() because no SSL running
	cc, err := grpc.Dial(c.AdminSvcUrl, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect:", err)
	}

	return pb.NewAdminServiceClient(cc)
}
