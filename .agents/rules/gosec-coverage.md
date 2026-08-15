# Go Security & Coverage Lessons (TIER 1) - AG Kit

> Lessons learned from CI failures regarding `gosec` and `go test` coverage.

---

## 🛡️ Gosec Rules (Securego)

When running `gosec` in CI, beware of the following false positives or strict rules:

1. **G115 (Integer Overflow Conversion):**
   - **Trigger:** Converting a larger integer type to a smaller one (e.g. `int64` -> `byte`, `int` -> `int32`).
   - **Fix:** If mathematically safe (e.g., `rand.Intn(10)` or validated config values), use `// #nosec G115` to suppress the warning instead of over-engineering type conversions.

2. **G101 (Hardcoded Credentials):**
   - **Trigger:** Variables named `messages` or maps with string keys sometimes trigger false positives if they contain words like "secret" or "password" in their dictionary (e.g., in i18n translation maps).
   - **Fix:** Annotate with `// #nosec G101` if it's purely a localization string map or non-sensitive data.

3. **G304, G302, G301 (File/Directory Permissions & Path Traversal):**
   - **Trigger:** Using `os.OpenFile` with variable paths or loose permissions.
   - **Fix:** 
     - Directories should be created with `0750` or tighter (not `0755`).
     - Files should be created with `0600` or tighter (not `0644`).
     - Ensure path is sanitized (e.g. `filepath.Clean`) and if a false positive remains, add `// #nosec G304`.

4. **G104 (Unhandled Errors):**
   - **Trigger:** `defer m.Close()` or ignoring errors without explicit discard.
   - **Fix:** Explicitly discard the error if safe: `_, _ = m.Close()` or `_ = f.Close()`. Do NOT use `//nolint:errcheck` because `gosec` ignores `golangci-lint` directives.

## 📊 Go Test Coverage

1. **Coverage drops to 0.0% unexpectedly:**
   - **Cause:** `go test -coverprofile=coverage.out ./...` defaults to only computing coverage for the package currently being tested. If your tests are separated into a `src/tests/...` directory, the coverage for your application logic `src/internal/...` will be 0%.
   - **Fix:** ALWAYS use the `-coverpkg=./...` flag when running tests across a project with separated test directories.
   - **Example:** `go test -v -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...`

2. **Unrealistic Initial Coverage Thresholds:**
   - **Cause:** Setting `COVERAGE_THRESHOLD: "70"` on an existing codebase with very few unit tests will block all CI/CD pipelines, because writing tests to jump from 12% to 70% in one PR is impractical.
   - **Fix:** Set the `COVERAGE_THRESHOLD` to a realistic baseline (e.g., `12%`) that the project currently meets.
   - **Continuous Improvement:** All *new code* must include tests. Incrementally raise the threshold as the project's overall test coverage grows, rather than enforcing a strict 70% Day 1 rule.

## 🐳 Docker Build & Go Version Mismatch

1. **`go mod download` fails in Docker with `go.mod requires go >= 1.X.Y (running go 1.X.Z)`:**
   - **Cause:** When updating the `go` directive in `go.mod` (e.g., to fix standard library vulnerabilities via `govulncheck`), the Docker image version in your `Dockerfile` and `Dockerfile.worker` must also be updated to match. Using a generic tag like `golang:1.25-alpine` can sometimes pull an outdated cached version, while using a strict tag like `golang:1.25.0-alpine` will break if `go.mod` requires `1.25.13`.
   - **Fix:** ALWAYS update the `FROM` statements in all Dockerfiles to explicitly use the exact patch version required by `go.mod` (e.g., `FROM golang:1.25.13-alpine AS builder` and `FROM golang:1.25.13-bookworm AS builder`).

2. **Docker build fails with `"/app/.env": not found`:**
   - **Cause:** CI environments often use `.dockerignore` to ignore `.env` files (to prevent leaking secrets). If a `Dockerfile` attempts to `COPY --from=builder /app/.env .env`, the build will fail because the file was never copied into the builder stage.
   - **Fix:** Do NOT bake `.env` files into Docker images. Delete any `COPY .env` instructions from Dockerfiles. Environment variables must be injected at runtime (e.g., via `docker run --env-file`, Kubernetes ConfigMaps, or GitHub Actions Secrets).
