package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"emc_lb/src/pkg/res"
)

type EcommerceHandler struct{}

func NewEcommerceHandler() *EcommerceHandler {
	return &EcommerceHandler{}
}

// 1. GET /api/v1/ecommerce/home
func (h *EcommerceHandler) GetHome(ctx *gin.Context) {
	data := gin.H{
		"banner": gin.H{
			"title": "Đánh thức vẻ đẹp tiềm ẩn",
			"description": "Khám phá bộ sưu tập thời trang thanh lịch, nhẹ nhàng và tôn vinh nét quyến rũ của bạn.",
			"image_url": "http://localhost:4566/ecommerce-images/hero.png",
		},
	}
	res.Success(ctx, http.StatusOK, data)
}

// 2. GET /api/v1/ecommerce/products
func (h *EcommerceHandler) GetProducts(ctx *gin.Context) {
	products := []gin.H{
		{
			"id": "prod-001",
			"name": "Váy Hồng Pastel Mùa Hè",
			"price": 550000,
			"description": "Chiếc váy nhẹ nhàng, thướt tha mang đậm phong cách nữ tính.",
			"image_url": "http://localhost:4566/ecommerce-images/dress.png",
		},
		{
			"id": "prod-002",
			"name": "Dây Chuyền Vàng Ngọc Trai",
			"price": 890000,
			"description": "Dây chuyền vàng sang trọng kết hợp ngọc trai tự nhiên.",
			"image_url": "http://localhost:4566/ecommerce-images/jewelry.png",
		},
	}
	res.Success(ctx, http.StatusOK, products)
}

// 3. GET /api/v1/ecommerce/products/:id
func (h *EcommerceHandler) GetProductDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	
	// Mock find product
	var product gin.H
	if id == "prod-001" {
		product = gin.H{
			"id": "prod-001",
			"name": "Váy Hồng Pastel Mùa Hè",
			"price": 550000,
			"description": "Chiếc váy nhẹ nhàng, thướt tha mang đậm phong cách nữ tính.",
			"image_url": "http://localhost:4566/ecommerce-images/dress.png",
			"in_stock": true,
		}
	} else {
		product = gin.H{
			"id": "prod-002",
			"name": "Dây Chuyền Vàng Ngọc Trai",
			"price": 890000,
			"description": "Dây chuyền vàng sang trọng kết hợp ngọc trai tự nhiên.",
			"image_url": "http://localhost:4566/ecommerce-images/jewelry.png",
			"in_stock": true,
		}
	}
	
	res.Success(ctx, http.StatusOK, product)
}

// 4. POST /api/v1/ecommerce/cart
type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

func (h *EcommerceHandler) AddToCart(ctx *gin.Context) {
	var req AddToCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, err)
		return
	}
	
	res.Success(ctx, http.StatusOK, gin.H{
		"message": "Đã thêm sản phẩm vào giỏ hàng thành công",
		"cart_item": req,
	})
}

// 5. POST /api/v1/ecommerce/checkout
type CheckoutRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Address  string `json:"address" binding:"required"`
}

func (h *EcommerceHandler) Checkout(ctx *gin.Context) {
	var req CheckoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, err)
		return
	}
	
	res.Success(ctx, http.StatusOK, gin.H{
		"message": "Đã đặt hàng thành công",
		"order_id": "ORD-12345",
		"shipping_info": req,
	})
}
