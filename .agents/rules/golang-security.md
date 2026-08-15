# Golang Security Rules (TIER 1) - AG Kit

> Always-active rules for Go security and vulnerability management.

---

## 🛡️ Vulnerability Management Protocol

**Whenever a vulnerability is detected (e.g., via `govulncheck`):**

1. **Standard Library Vulnerabilities:**
   - Standard library bugs (e.g., `net/url`, `html/template`, `crypto/tls`) are fixed in Go patch releases.
   - **Action:** Update the `go` directive in `go.mod` (e.g., `go 1.25.13`) AND update the CI workflow Go version (e.g., `GO_VERSION: "1.25.13"` in `.github/workflows/ci.yml`).

2. **Third-Party Module Vulnerabilities:**
   - Examples: `github.com/golang-jwt/jwt`, `github.com/aws/aws-sdk-go-v2`.
   - **Action:** Use `go get module@version` to explicitly upgrade to the patched version, then run `go mod tidy`.

3. **Prevention:**
   - Always verify dependencies against the Go Vulnerability Database (`govulncheck ./...`).
   - Do not ignore `-show verbose` warnings if they indicate active traces in our code.
