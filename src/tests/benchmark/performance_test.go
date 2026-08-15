package benchmark

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupMockRouter creates a basic Gin router with a mocked endpoint for benchmarking throughput
func setupMockRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Mock ListProducts endpoint
	r.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": []map[string]interface{}{
				{"id": "prod_1", "name": "Laptop", "price": 1000},
				{"id": "prod_2", "name": "Phone", "price": 500},
			},
		})
	})

	// Mock CreateOrder endpoint
	r.POST("/api/v1/orders", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
		// Simulate fast processing
		c.JSON(http.StatusCreated, gin.H{
			"status":   "success",
			"order_id": "order_uuid_123",
		})
	})

	return r
}

// BenchmarkAPIThroughput tests the max requests/sec for a GET endpoint
func BenchmarkAPIThroughput(b *testing.B) {
	router := setupMockRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/products", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkCreateOrder tests the throughput of a POST endpoint processing JSON
func BenchmarkCreateOrder(b *testing.B) {
	router := setupMockRouter()

	payload := map[string]interface{}{
		"shop_id": "shop_1",
		"items": []map[string]interface{}{
			{"product_id": "prod_1", "quantity": 1},
		},
	}
	body, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
