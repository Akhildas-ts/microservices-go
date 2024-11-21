package services

import (
	"context"
	"net/http"

	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/client"
	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/db"
	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/models"
	"github.com/Akhildas-ts/go-grpc-cart-svc/pkg/pb"
)

type Server struct {
	H          db.Handler
	ProductSvc client.ProductServiceClient
	pb.UnimplementedCartServer
}

func (s *Server) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.AddToCartResponse, error) {
	// Step 1: Validate the product
	product, err := s.ProductSvc.FindOne(req.ProductID)
	if err != nil {
		return &pb.AddToCartResponse{Error: err.Error()}, nil
	} else if product.Status >= http.StatusNotFound {
		return &pb.AddToCartResponse{Error: product.Error}, nil
	} else if product.Data.Stock < req.Quantity {
		return &pb.AddToCartResponse{Error: "Stock too less"}, nil
	}

	// Step 2: Add item to the cart
	addCart := models.Cart{
		ProductID: req.ProductID,
		UserId:    req.UserID,
		Quantity:  req.Quantity,
	}
	if result := s.H.DB.Create(&addCart); result.Error != nil {
		return &pb.AddToCartResponse{Error: result.Error.Error()}, nil
	}

	// Step 3: Fetch all cart items for the user
	var cartItems []models.Cart
	if result := s.H.DB.Where("user_id = ?", req.UserID).Find(&cartItems); result.Error != nil {
		return &pb.AddToCartResponse{Error: result.Error.Error()}, nil
	}

	// Step 4: Prepare the response items and calculate total amount
	var items []*pb.CartDetails

	var totalAmount float32

	for _, item := range cartItems {
		product, err := s.ProductSvc.FindOne(item.ProductID)
		if err != nil {
			return &pb.AddToCartResponse{Error: err.Error()}, nil
		}
		totalAmount += float32(item.Quantity) * float32(product.Data.Price)

		items = append(items, &pb.CartDetails{
			ProductID:  item.ProductID,
			Quantity:   float32(item.Quantity),
			TotalPrice: float32(product.Data.Price),
		})
	}

	// Step 5: Return the updated cart and total amount
	return &pb.AddToCartResponse{
		Cart:  items,
		Price: totalAmount,
	}, nil
}

func (s *Server) GetAllItemsFromCart(ctx context.Context, req *pb.GetAllItemsFromCartRequest) (*pb.GetAllItemsFromCartResponse, error) {
	var cartItems []models.Cart

	// Fetch all cart items for the user
	if result := s.H.DB.Where("user_id = ?", req.UserID).Find(&cartItems); result.Error != nil {
		return &pb.GetAllItemsFromCartResponse{Error: result.Error.Error()}, nil
	}

	// Prepare the response items
	var items []*pb.CartDetails
	for _, item := range cartItems {
		items = append(items, &pb.CartDetails{
			ProductID: item.ProductID,
			Quantity:  float32(item.Quantity),
		})
	}

	return &pb.GetAllItemsFromCartResponse{
		Cart: items,
	}, nil
}

func (s *Server) TotalAmountInCart(ctx context.Context, req *pb.TotalAmountInCartRequest) (*pb.TotalAmountInCartResponse, error) {
	var cartItems []models.Cart

	// Fetch all cart items for the user
	if result := s.H.DB.Where("user_id = ?", req.UserID).Find(&cartItems); result.Error != nil {
		return &pb.TotalAmountInCartResponse{Error: result.Error.Error()}, nil
	}

	// Calculate the total amount
	var totalAmount float32
	for _, item := range cartItems {
		product, err := s.ProductSvc.FindOne(item.ProductID)
		if err != nil {
			return &pb.TotalAmountInCartResponse{Error: err.Error()}, nil
		}
		totalAmount += float32(item.Quantity) * float32(product.Data.Price)
	}

	return &pb.TotalAmountInCartResponse{
		Data: totalAmount,
	}, nil
}
