# EMC LB

Scaffold ban đầu cho dự án Go sử dụng Postgres và MongoDB.

## Tính năng

- Postgres được dựng bằng `docker compose`
- MongoDB được dựng bằng `docker compose`
- Sử dụng `sqlc` để tạo mã truy vấn Go từ SQL
- Sử dụng `air` để tự động reload khi phát triển
- Mã nguồn đặt trong `src/`

## Khởi động nhanh

1. `make docker-up`
2. `make sqlc` (sau khi cài sqlc)
3. `make dev`

## Biến môi trường

Các setting chính nằm trong `.env`.

## Cấu trúc chính

- `src/cmd/server/main.go`
- `src/internal/config`
- `src/internal/db`
- `src/internal/handlers`
- `src/db/schema.sql`
- `src/db/queries.sql`
