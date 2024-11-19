package routes

import (
	"context"
	"net/http"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/cart/pb"
	"github.com/gin-gonic/gin"
)

type GetCartRequest struct {
	Userid int64 `json:"user_id"`
}

func GetCart(ctx *gin.Context, c pb.CartClient) {
	b := GetCartRequest{}

	if err := ctx.BindJSON(&b); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	res, err := c.GetCart(context.Background(), &pb.GetCartRequest{
		UserID: b.Userid,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	ctx.JSON(http.StatusCreated, &res)
}
