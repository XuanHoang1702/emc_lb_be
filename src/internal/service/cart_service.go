package service

import (
	"context"
	"fmt"
	"net/http"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
)

type CartService interface {
	GetCart(ctx context.Context, userID string) (entities.CartResponse, error)
	AddItem(ctx context.Context, userID string, req entities.AddCartItemRequest) (entities.CartResponse, error)
	UpdateItem(ctx context.Context, userID string, productID string, req entities.UpdateCartItemRequest) (entities.CartResponse, error)
	RemoveItem(ctx context.Context, userID string, productID string) (entities.CartResponse, error)
	ClearCart(ctx context.Context, userID string) error
}

type cartService struct {
	cartRepository    repository.CartRepository
	productRepository repository.ProductRepository
}

func NewCartService(cartRepository repository.CartRepository, productRepository repository.ProductRepository) CartService {
	return &cartService{
		cartRepository:    cartRepository,
		productRepository: productRepository,
	}
}

func (s *cartService) GetCart(ctx context.Context, userID string) (entities.CartResponse, error) {
	cart, err := s.cartRepository.GetByUserID(ctx, userID)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to get cart", erres.CommonInternal)
	}

	return s.buildCartResponse(ctx, cart)
}

func (s *cartService) AddItem(ctx context.Context, userID string, req entities.AddCartItemRequest) (entities.CartResponse, error) {
	// Verify product exists and has enough stock
	product, err := s.productRepository.GetByID(ctx, req.ProductID)
	if err != nil {
		return entities.CartResponse{}, &res.AppError{
			Message:    "Product not found",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	cart, err := s.cartRepository.GetByUserID(ctx, userID)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to get cart", erres.CommonInternal)
	}

	// Check if item already exists in cart
	itemExists := false
	var currentQty int64 = 0
	
	for i, item := range cart.Items {
		if item.ProductID == req.ProductID {
			itemExists = true
			currentQty = item.Quantity
			cart.Items[i].Quantity += req.Quantity
			break
		}
	}

	if !product.AllowBackorder && product.Stock < (currentQty + req.Quantity) {
		return entities.CartResponse{}, &res.AppError{
			Message:    fmt.Sprintf("Not enough stock. Only %d available.", product.Stock),
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if !itemExists {
		cart.Items = append(cart.Items, entities.CartItem{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		})
	}

	savedCart, err := s.cartRepository.Save(ctx, cart)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to save cart", erres.CommonInternal)
	}

	return s.buildCartResponse(ctx, savedCart)
}

func (s *cartService) UpdateItem(ctx context.Context, userID string, productID string, req entities.UpdateCartItemRequest) (entities.CartResponse, error) {
	cart, err := s.cartRepository.GetByUserID(ctx, userID)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to get cart", erres.CommonInternal)
	}

	itemExists := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			itemExists = true
			
			// Verify stock
			product, pErr := s.productRepository.GetByID(ctx, productID)
			if pErr == nil && !product.AllowBackorder && product.Stock < req.Quantity {
				return entities.CartResponse{}, &res.AppError{
					Message:    fmt.Sprintf("Not enough stock. Only %d available.", product.Stock),
					Code:       erres.CommonBadRequest,
					StatusCode: http.StatusBadRequest,
				}
			}
			
			cart.Items[i].Quantity = req.Quantity
			break
		}
	}

	if !itemExists {
		return entities.CartResponse{}, &res.AppError{
			Message:    "Item not found in cart",
			Code:       erres.CommonNotFound,
			StatusCode: http.StatusNotFound,
		}
	}

	savedCart, err := s.cartRepository.Save(ctx, cart)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to update cart", erres.CommonInternal)
	}

	return s.buildCartResponse(ctx, savedCart)
}

func (s *cartService) RemoveItem(ctx context.Context, userID string, productID string) (entities.CartResponse, error) {
	cart, err := s.cartRepository.GetByUserID(ctx, userID)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to get cart", erres.CommonInternal)
	}

	var newItems []entities.CartItem
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}

	cart.Items = newItems

	savedCart, err := s.cartRepository.Save(ctx, cart)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to update cart", erres.CommonInternal)
	}

	return s.buildCartResponse(ctx, savedCart)
}

func (s *cartService) ClearCart(ctx context.Context, userID string) error {
	return s.cartRepository.ClearCart(ctx, userID)
}

func (s *cartService) buildCartResponse(ctx context.Context, cart entities.Cart) (entities.CartResponse, error) {
	response := entities.CartResponse{
		ID:          cart.ID,
		UserID:      cart.UserID,
		Items:       []entities.CartItemResponse{},
		TotalAmount: 0,
		CreatedAt:   cart.CreatedAt,
		UpdatedAt:   cart.UpdatedAt,
	}

	for _, item := range cart.Items {
		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			// If product was deleted, we still include it but mark it invalid or just skip.
			// For this implementation, let's skip invalid items from response but maybe we shouldn't fail.
			continue
		}

		subTotal := product.Price * float64(item.Quantity)
		response.TotalAmount += subTotal

		response.Items = append(response.Items, entities.CartItemResponse{
			ProductID:    product.ID,
			ProductName:  product.Name,
			ProductImage: product.Thumbnail,
			Price:        product.Price,
			Quantity:     item.Quantity,
			SubTotal:     subTotal,
		})
	}

	return response, nil
}
