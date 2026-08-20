# Load Testing (k6)

Mô phỏng traffic thực tế qua **nginx load balancer** (không phải micro-benchmark),
phản ánh đúng hành vi hệ thống khi có hàng nghìn request đồng thời như dịp
Black Friday / sale lớn.

## Chạy locally

Yêu cầu: docker compose stack đang chạy (`make docker-up`), nginx lắng nghe trên `http://localhost:80`.

Trước khi chạy, seed dữ liệu cần thiết (roles/quyền + user đã xác thực email +
sản phẩm thật) vì luồng login → giỏ → checkout yêu cầu **user email_verified**
và **product có id MongoDB hợp lệ**:

```bash
export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/emc_lb?sslmode=disable"
export SYSTEM_SECRET="<giống SYSTEM_SECRET trong .env của stack>"

# Roles/quyền (lỗi re-run là bình thường)
go run ./src/cmd/seed || true

# 200 user đã xác thực email, mật khẩu LoadTest123!
LOADTEST_USERS=200 go run ./src/cmd/loadtest-seed

# Sản phẩm thật trong MongoDB (id hex) — dùng mongosh, tham khảo bước
# "Seed load-test products" trong .github/workflows/load-test.yml
```

```bash
# Nếu chưa cài k6: 
#   brew install k6          (macOS)
#   choco install k6         (Windows)
#   snap install k6          (Ubuntu)

# Chạy toàn bộ kịch bản load
PRODUCT_IDS="<hex-id-sản-phẩm>," LOADTEST_USER_COUNT=200 k6 run loadtest/checkout-flow.js

# Chạy riêng một scenario
k6 run --only smoke loadtest/checkout-flow.js
k6 run --only ramp-up loadtest/checkout-flow.js
k6 run --only soak loadtest/checkout-flow.js
k6 run --only spike loadtest/checkout-flow.js
```

## Chạy trong CI

Workflow `.github/workflows/load-test.yml` tự động:
1. Bật toàn bộ stack (postgres, mongo, redis, localstack, 3 API nodes, nginx)
2. Chờ các service healthy
3. Apply migrations + seed dữ liệu
4. Chạy k6 theo 4 scenario
5. Fail nếu vượt ngưỡng SLO (error rate, p95 latency, throughput)

## Kịch bản

| Scenario | Mô tả | VU / Duration | Ngưỡng |
|----------|-------|---------------|--------|
| smoke | Kiểm tra pipeline, 1 VU | 1 VU / 30s | p95 < 1000ms, errors = 0 |
| ramp-up | Tăng dần tới đỉnh như giờ cao điểm | 0 → 500 VU / 5m | p95 < 800ms, errors < 1% |
| soak | Kiểm tra rò rỉ bộ nhớ / cạn kết nối | 100 VU / 10m | p95 < 800ms, errors < 1% |
| spike | Sốc traffic như flash sale | 0 → 2000 VU / 1m | errors < 5% |

## Luồng nghiệp vụ mô phỏng (`checkout-flow.js`)

```
1. GET  /health                     → kiểm tra LB lên
2. GET  /api/v1/products           → duyệt danh sách sản phẩm
3. GET  /api/v1/products/:id       → xem chi tiết sản phẩm
4. POST /api/v1/user/login         → đăng nhập user đã seed (email_verified)
5. POST /api/v1/cart/items         → thêm vào giỏ hàng
6. POST /api/v1/orders             → checkout tạo đơn
```

Mỗi VU chạy một luồng hoàn chỉnh, bắt chước hành vi người mua thực tế.
User login được chọn xoay vòng trong pool `loadtest_001..N@example.com` (do
`loadtest-seed` tạo) — mật khẩu chung `LoadTest123!`.

## Cấu hình

- `BASE_URL`: địa chỉ load balancer. Mặc định `http://localhost`. Trong CI tự set
  tới nginx container.
- `API_KEY`: dùng cho `/health/detail`. Mặc định lấy từ biến môi trường
  `SYSTEM_SECRET` của stack, CI truyền qua.
- `PRODUCT_IDS`: danh sách product id (MongoDB hex) phân tách bằng dấu phẩy,
  dùng cho thêm giỏ/checkout. Trong CI lấy từ bước seed MongoDB.
- `LOADTEST_USER_COUNT`: số user đã seed để login. Mặc định 10; CI seed 200.