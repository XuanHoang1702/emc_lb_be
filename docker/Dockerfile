# Builder stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go env -w GO111MODULE=on && go mod download

# Copy source
COPY . .

# Build binary with optimized flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /usr/local/bin/emc_lb ./src/cmd/server

# Final stage
FROM gcr.io/distroless/static-debian12

COPY --from=builder /usr/local/bin/emc_lb /usr/local/bin/emc_lb
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/emc_lb"]
