package services

import (
	"context"
	"net/http"

	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/db"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/models"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/pb"
	"github.com/Akhildas-ts/go-grpc-admin-svc/pkg/utils"
)

type Server struct {
	H   db.Handler
	Jwt utils.JwtWrapper
	pb.UnimplementedAdminServiceServer
}

func (s *Server) SignupAdmin(ctx context.Context, req *pb.SignupAdminRequest) (*pb.SignupAdminResponse, error) {

	var admin models.User
	if result := s.H.DB.Where(&models.User{Email: req.Email}).First(&admin); result.Error == nil {
		return &pb.SignupAdminResponse{
			Status: http.StatusConflict,
			Error:  "get issue from finding user",
		}, nil

	}

	admin.Email = req.Email
	admin.Password = utils.HashPassword(req.Password)
	admin.Isadmin = true

	s.H.DB.Create(&admin)

	return &pb.SignupAdminResponse{
		Status: http.StatusCreated,
		Error:  "admin created succefully",
	}, nil

}

func (s *Server) LoginAdmin(ctx context.Context, req *pb.LoginAdminRequest) (*pb.LoginAdminResponse, error) {
	var user models.User

	// Find user by email
	if result := s.H.DB.Where(&models.User{Email: req.Email}).First(&user); result.Error != nil {
		return &pb.LoginAdminResponse{
			Status:  http.StatusNotFound,
			Message: "not found any user",
		}, nil
	}

	// Check if password matches
	match := utils.CheckPasswordHash(req.Password, user.Password)
	if !match {
		return &pb.LoginAdminResponse{
			Status:  http.StatusNotFound,
			Message: "password not match ",
		}, nil
	}

	// Check if user is admin
	if !user.Isadmin {
		return &pb.LoginAdminResponse{
			Status:  http.StatusUnauthorized,
			Message: "user found in this account sorry ..",
		}, nil
	}

	// Generate JWT token
	token, err := s.Jwt.GenerateToken(user)
	if err != nil {
		return &pb.LoginAdminResponse{
			Status:  http.StatusInternalServerError,
			Message: "error from generating token",
		}, nil
	}

	return &pb.LoginAdminResponse{
		Status: http.StatusOK,
		Token:  token,
	}, nil
}
