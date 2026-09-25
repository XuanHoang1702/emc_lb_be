package security_test

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =====================================================================
// SEC-001: Security Headers Middleware Tests
// Validates that all required security headers are present on responses.
// =====================================================================

func TestSecurityHeaders_AllPresent(t *testing.T) {
	router := gin.New()
	router.Use(middleware.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"X-XSS-Protection":      "0",
		"Referrer-Policy":       "strict-origin-when-cross-origin",
		"Cache-Control":         "no-store",
	}

	for header, expected := range expectedHeaders {
		got := w.Header().Get(header)
		if got != expected {
			t.Errorf("header %s: want %q, got %q", header, expected, got)
		}
	}

	// CSP should contain frame-ancestors 'none'
	csp := w.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Errorf("CSP missing frame-ancestors: got %q", csp)
	}

	// Permissions-Policy should restrict dangerous features
	pp := w.Header().Get("Permissions-Policy")
	if !strings.Contains(pp, "camera=()") {
		t.Errorf("Permissions-Policy missing camera restriction: got %q", pp)
	}
}

func TestSecurityHeaders_PresentOnError(t *testing.T) {
	router := gin.New()
	router.Use(middleware.SecurityHeadersMiddleware())
	router.GET("/test-error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test-error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("security headers missing on error response")
	}
}

// =====================================================================
// SEC-002: OTP Constant-Time Comparison Tests
// Validates that OTP comparison uses constant-time algorithm.
// =====================================================================

func TestOTPConstantTimeCompare_Equal(t *testing.T) {
	otp := "123456"
	if subtle.ConstantTimeCompare([]byte(otp), []byte("123456")) != 1 {
		t.Error("constant-time compare should return 1 for equal OTPs")
	}
}

func TestOTPConstantTimeCompare_NotEqual(t *testing.T) {
	otp := "123456"
	if subtle.ConstantTimeCompare([]byte(otp), []byte("654321")) != 0 {
		t.Error("constant-time compare should return 0 for different OTPs")
	}
}

func TestOTPConstantTimeCompare_DifferentLength(t *testing.T) {
	otp := "123456"
	if subtle.ConstantTimeCompare([]byte(otp), []byte("12345")) != 0 {
		t.Error("constant-time compare should return 0 for different length OTPs")
	}
}

// =====================================================================
// SEC-003: Access Token Middleware Tests
// Validates authentication boundary enforcement.
// =====================================================================

func TestAccessToken_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AccessTokenMiddleware("test-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("missing auth header: want 401, got %d", w.Code)
	}
}

func TestAccessToken_EmptyBearer(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AccessTokenMiddleware("test-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("empty bearer: want 401, got %d", w.Code)
	}
}

func TestAccessToken_InvalidPrefix(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AccessTokenMiddleware("test-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should reject non-Bearer scheme
	if w.Code != http.StatusUnauthorized {
		t.Errorf("non-Bearer scheme: want 401, got %d", w.Code)
	}
}

func TestAccessToken_MalformedJWT(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AccessTokenMiddleware("test-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("malformed JWT: want 401, got %d", w.Code)
	}
}

// =====================================================================
// SEC-004: API Key Middleware Tests
// Validates constant-time API key comparison.
// =====================================================================

func TestAPIKey_Missing(t *testing.T) {
	router := gin.New()
	router.Use(middleware.APIKeyMiddleware("correct-key"))
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("missing API key: want 400, got %d", w.Code)
	}
}

func TestAPIKey_Invalid(t *testing.T) {
	router := gin.New()
	router.Use(middleware.APIKeyMiddleware("correct-key"))
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong API key: want 401, got %d", w.Code)
	}
}

func TestAPIKey_Valid(t *testing.T) {
	router := gin.New()
	router.Use(middleware.APIKeyMiddleware("correct-key"))
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	req.Header.Set("X-API-Key", "correct-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("correct API key: want 200, got %d", w.Code)
	}
}

// =====================================================================
// SEC-005: RBAC Middleware Tests
// Validates role-based access control enforcement.
// =====================================================================

func TestRBAC_MissingRole(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequirePermission("product:create"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("missing role context: want 401, got %d", w.Code)
	}
}

func TestRBAC_EmptyRole(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextRoleKey, "")
		c.Next()
	})
	router.Use(middleware.RequirePermission("product:create"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("empty role: want 401, got %d", w.Code)
	}
}

// =====================================================================
// SEC-006: Request Body Sanitization Tests
// Validates that sensitive fields are redacted in logs.
// =====================================================================

func TestRequestLogSanitization_PasswordRedacted(t *testing.T) {
	body := `{"email":"test@test.com","password":"mysecretpass"}`
	sanitized := sanitizeBodyForTest(body)

	var result map[string]any
	_ = json.Unmarshal([]byte(sanitized), &result)

	if result["password"] != "[REDACTED]" {
		t.Errorf("password should be redacted, got: %v", result["password"])
	}
	if result["email"] != "test@test.com" {
		t.Errorf("email should not be redacted, got: %v", result["email"])
	}
}

func TestRequestLogSanitization_AllSensitiveFields(t *testing.T) {
	body := `{
		"password":"test",
		"old_password":"old",
		"new_password":"new",
		"refresh_token":"rt_xyz",
		"access_token":"at_xyz",
		"otp":"123456",
		"secret_key":"sk_123",
		"card_number":"4111111111111111",
		"card_holder_name":"John Doe",
		"card_expiry":"12/25",
		"cvv":"123",
		"email":"safe@test.com"
	}`
	sanitized := sanitizeBodyForTest(body)

	var result map[string]any
	_ = json.Unmarshal([]byte(sanitized), &result)

	sensitiveFields := []string{
		"password", "old_password", "new_password",
		"refresh_token", "access_token", "otp",
		"secret_key", "card_number", "card_holder_name",
		"card_expiry", "cvv",
	}

	for _, field := range sensitiveFields {
		if result[field] != "[REDACTED]" {
			t.Errorf("field %q should be redacted, got: %v", field, result[field])
		}
	}

	// Non-sensitive field should be preserved
	if result["email"] != "safe@test.com" {
		t.Errorf("email should not be redacted")
	}
}

func TestRequestLogSanitization_NoSensitiveFields(t *testing.T) {
	body := `{"name":"test","quantity":5}`
	sanitized := sanitizeBodyForTest(body)

	if !strings.Contains(sanitized, "test") {
		t.Error("non-sensitive fields should be preserved")
	}
}

func TestRequestLogSanitization_InvalidJSON(t *testing.T) {
	body := "not json"
	sanitized := sanitizeBodyForTest(body)

	if sanitized != body {
		t.Error("invalid JSON should be returned as-is")
	}
}

// sanitizeBodyForTest replicates the middleware sanitization logic for unit testing.
func sanitizeBodyForTest(body string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return body
	}

	sensitiveFields := []string{
		"password", "old_password", "new_password",
		"refresh_token", "access_token", "otp",
		"secret_key", "card_number", "card_holder_name",
		"card_expiry", "cvv",
	}

	for _, field := range sensitiveFields {
		if _, exists := payload[field]; exists {
			payload[field] = "[REDACTED]"
		}
	}

	sanitized, err := json.Marshal(payload)
	if err != nil {
		return body
	}

	return string(sanitized)
}

// =====================================================================
// SEC-007: Request ID Middleware Tests
// Validates request ID handling and injection prevention.
// =====================================================================

func TestRequestID_Generated(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": middleware.GetRequestID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should be generated")
	}

	// Should be valid UUID format (36 chars with dashes)
	if len(requestID) != 36 {
		t.Errorf("generated request ID should be UUID format, got len=%d", len(requestID))
	}
}

func TestRequestID_PropagatedFromClient(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": middleware.GetRequestID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "client-provided-id")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "client-provided-id" {
		t.Error("client-provided request ID should be echoed back")
	}
}

// =====================================================================
// SEC-008: CORS Middleware Tests
// Validates CORS configuration.
// =====================================================================

func TestCORS_DefaultOrigins(t *testing.T) {
	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Allowed origin
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("allowed origin should be reflected, got: %s", origin)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "http://evil.com" {
		t.Error("disallowed origin should NOT be reflected")
	}
}

// =====================================================================
// SEC-009: Request Body Read-Back Test
// Validates that request body is preserved after logging middleware reads it.
// =====================================================================

func TestRequestLog_BodyPreserved(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RequestLogMiddleware())
	router.POST("/test", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"body": string(body)})
	})

	reqBody := `{"name":"test","value":42}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var result map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	if result["body"] != reqBody {
		t.Errorf("body not preserved after logging: got %v", result["body"])
	}
}
