package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersMiddleware adds defense-in-depth HTTP security headers to all responses.
//
// Headers set:
//   - X-Content-Type-Options: nosniff    → Prevents MIME-type sniffing attacks
//   - X-Frame-Options: DENY              → Prevents clickjacking via iframe embedding
//   - X-XSS-Protection: 0               → Disables legacy XSS auditor (CSP preferred)
//   - Referrer-Policy: strict-origin-when-cross-origin → Limits referrer leakage
//   - Permissions-Policy: ...            → Disables unnecessary browser features
//   - Content-Security-Policy: default-src 'self' → Restricts resource loading origins
//   - Cache-Control: no-store            → Prevents caching of API responses
//
// Note: Strict-Transport-Security (HSTS) should be set at the reverse-proxy/LB level,
// not in the app, to avoid issues in local dev environments (HTTP-only).
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "DENY")
		ctx.Header("X-XSS-Protection", "0")
		ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		ctx.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		ctx.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		ctx.Header("Cache-Control", "no-store")
		ctx.Next()
	}
}
