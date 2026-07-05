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
	ApplyCoupon(ctx context.Context, userID string, req entities.ApplyCouponRequest) (entities.CartResponse, error)
}

type cartService struct {
	cartRepository    repository.CartRepository
	productRepository repository.ProductRepository
	couponService     CouponService
}

func NewCartService(cartRepository repository.CartRepository, productRepository repository.ProductRepository, couponService CouponService) CartService {
	return &cartService{
		cartRepository:    cartRepository,
		productRepository: productRepository,
		couponService:     couponService,
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

func (s *cartService) ApplyCoupon(ctx context.Context, userID string, req entities.ApplyCouponRequest) (entities.CartResponse, error) {
	cart, err := s.cartRepository.GetByUserID(ctx, userID)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to get cart", erres.CommonInternal)
	}

	// Just save the coupon code to the cart. Validation happens in buildCartResponse
	cart.CouponCode = req.CouponCode
	savedCart, err := s.cartRepository.Save(ctx, cart)
	if err != nil {
		return entities.CartResponse{}, res.WrapError(err, "Failed to update cart", erres.CommonInternal)
	}

	return s.buildCartResponse(ctx, savedCart)
}

func (s *cartService) buildCartResponse(ctx context.Context, cart entities.Cart) (entities.CartResponse, error) {
	response := entities.CartResponse{
		ID:             cart.ID,
		UserID:         cart.UserID,
		Items:          []entities.CartItemResponse{},
		CouponCode:     cart.CouponCode,
		SubTotal:       0,
		DiscountAmount: 0,
		TaxAmount:      0,
		TotalAmount:    0,
		CreatedAt:      cart.CreatedAt,
		UpdatedAt:      cart.UpdatedAt,
	}

	for _, item := range cart.Items {
		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			// Skip invalid/deleted products
			continue
		}

		subTotal := product.Price * float64(item.Quantity)
		response.SubTotal += subTotal

		response.Items = append(response.Items, entities.CartItemResponse{
			ProductID:    product.ID,
			ProductName:  product.Name,
			ProductImage: product.Thumbnail,
			Price:        product.Price,
			Quantity:     item.Quantity,
			SubTotal:     subTotal,
		})
	}

	// Calculate discount
	if cart.CouponCode != "" {
		coupon, err := s.couponService.ValidateCouponForAmount(ctx, cart.CouponCode, response.SubTotal)
		if err == nil {
			if coupon.Type == "percentage" {
				discount := response.SubTotal * (coupon.Value / 100.0)
				if coupon.MaxDiscountAmount > 0 && discount > coupon.MaxDiscountAmount {
					discount = coupon.MaxDiscountAmount
				}
				response.DiscountAmount = discount
			} else if coupon.Type == "fixed_amount" {
				response.DiscountAmount = coupon.Value
				if response.DiscountAmount > response.SubTotal {
					response.DiscountAmount = response.SubTotal
				}
			}
		}
	}

	// Calculate Tax and Total
	amountAfterDiscount := response.SubTotal - response.DiscountAmount
	response.TaxAmount = amountAfterDiscount * 0.10 // 10% VAT
	response.TotalAmount = amountAfterDiscount + response.TaxAmount

	return response, nil
}
