package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Akhildas_ts/go-grpc-api-gateway/pkg/admin/pb"
	"github.com/gin-gonic/gin"
)

type SignUpRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SingupAdmin(ctx *gin.Context, c pb.AdminServiceClient) {
	body := SignUpRequestBody{}

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	res, err := c.SignupAdmin(context.Background(), &pb.SignupAdminRequest{
		Email:    body.Email,
		Password: body.Password,
	})

	if err != nil {
		fmt.Println("Error from singupamdin connection to the grpc....", err)
		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	ctx.JSON(http.StatusAccepted, &res)
}
