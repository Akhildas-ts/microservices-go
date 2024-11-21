package cart

import (
	"fmt"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/cart/pb"
	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/config"
	"google.golang.org/grpc"
)

type ServiceClient struct {
	Client pb.CartClient
}

func InitServiceClient(c *config.Config) pb.CartClient {
	cc, err := grpc.Dial(c.CartSvcUrl, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Count not connect ", err)
	}

	return pb.NewCartClient(cc)
}
