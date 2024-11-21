package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/cart/pb"
	"github.com/gin-gonic/gin"
)

type AddCartRequest struct {
	Product_id int64 `json:"product_id"`
	User_id    int64 `json:"user_id"`
	Quantity   int64 `json"quantity"`
}

func AddToCart(ctx *gin.Context, c pb.CartClient) {
	b := AddCartRequest{}

	if err := ctx.BindJSON(&b); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	res, err := c.AddToCart(context.Background(), &pb.AddToCartRequest{
		ProductID: b.Product_id,
		UserID:    b.User_id,
		Quantity:  b.Quantity,
	})

	if err != nil {

		fmt.Println("Error from connection of grpc .. ")
		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	ctx.JSON(http.StatusCreated, &res)
}
