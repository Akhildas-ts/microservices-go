package routes

import (
	"context"
	"net/http"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin/pb"
	"github.com/gin-gonic/gin"
)

type LoginRequestBody struct {
	Email    string `json: "email"`
	Password string `json: "password"`
}

func LoginAdmin(ctx *gin.Context, c pb.AdminServiceClient) {

	b := LoginRequestBody{}

	if err := ctx.BindJSON(&b); err != nil {

		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res, err := c.LoginAdmin(context.Background(), &pb.LoginAdminRequest{
		Email:    b.Email,
		Password: b.Password,
	})

	if err != nil {

		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	ctx.JSON(http.StatusAccepted, res)
}
