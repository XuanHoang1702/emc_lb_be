#!/usr/bin/env bash
#
# E2E nghiệp vụ thật qua toàn bộ stack (nginx LB -> API -> postgres/mongo/redis):
#   Đăng ký -> Xác thực OTP -> Đăng nhập -> Danh sách SP -> Thêm vào giỏ
#   -> Xem giỏ -> Tạo đơn hàng (checkout) -> Xem đơn hàng của tôi
#
# Cách chạy:
#   bash e2e/e2e-test.sh
#   ECC_ENV_FILE_ABS=.env.ci ECC_ENV_FILE=../.env.ci API_KEY=... bash e2e/e2e-test.sh

set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker/docker-compose.yml}"
# --env-file truyền cho docker compose tính theo cwd (repo root).
# env_file: ${ECC_ENV_FILE:-...} trong compose tính theo thư mục docker/.
ENV_FILE="${ECC_ENV_FILE_ABS:-.env.development}"
ECC_ENV_FILE="${ECC_ENV_FILE:-../.env.development}"
BASE_URL="${BASE_URL:-http://localhost}"
MONGO_USER="${MONGO_USER:-root}"
MONGO_PASS="${MONGO_PASS:-mongo_pass}"
REDIS_PASS="${REDIS_PASS:-redis_pass}"

PASS=0
FAIL=0

log() { echo -e "\n\033[1;35m[e2e]\033[0m $*"; }
ok()  { echo -e "  \033[1;32m✔\033[0m $*"; PASS=$((PASS+1)); }
bad() { echo -e "  \033[1;31m✘\033[0m $*"; FAIL=$((FAIL+1)); }

compose() { ECC_ENV_FILE="$ECC_ENV_FILE" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"; }

wait_healthy() {
  # 1) /health up (nginx -> API reachable)
  # 2) một endpoint nghiệp vụ trả về kết quả ổn định (không còn 502/503/000
  #    trong lúc node đang khởi động). /products gọi qua postgres/mongo.
  local i
  for i in $(seq 1 60); do
    if curl -sf "$BASE_URL/health" 2>/dev/null | grep -q '"status":"up"' &&
       ! curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$BASE_URL/api/v1/products" \
         | grep -qE '^(000|502|503)$'; then
      return 0
    fi
    sleep 2
  done
  return 1
}

api() {
  # api METHOD PATH [JSON_BODY] [TOKEN]
  local method="$1" path="$2" body="${3:-}" token="${4:-}"
  local args=(-s -w $'\n%{http_code}' -X "$method" -H "Content-Type: application/json")
  if [ -n "$token" ]; then args+=(-H "Authorization: Bearer $token"); fi
  if [ -n "$body" ]; then args+=(-d "$body"); fi
  curl "${args[@]}" "$BASE_URL$path"
}

body_of()  { echo "$1" | head -1; }
code_of()  { echo "$1" | tail -1; }
json_get() { echo "$1" | grep -o "$2" | head -1 | cut -d'"' -f4; }

# ------------------------------------------------------------------
log "Khởi động stack (nếu chưa chạy)..."
compose up -d --build >/dev/null
if ! wait_healthy; then
  bad "stack không healthy sau 180s"
  compose logs || true
  exit 1
fi
ok "stack healthy"
# ------------------------------------------------------------------
log "Seed dữ liệu (roles/permissions)..."
go run ./src/cmd/seed >/dev/null 2>&1 || true
ok "seed hoàn tất (re-run trên DB có dữ liệu là bình thường)"

# ------------------------------------------------------------------
log "Seed 1 sản phẩm test vào MongoDB..."
MONGO_DB_NAME="${MONGO_DB:-emc_lb}"
PRODUCT_ID=$(compose exec -T mongo mongosh --quiet \
  --username "$MONGO_USER" --password "$MONGO_PASS" --authenticationDatabase admin \
  --eval "
const dbc = db.getSiblingDB('$MONGO_DB_NAME');
const existing = dbc.products.findOne({ slug: 'e2e-test-product' });
if (existing) { print(existing._id.toString()); } else {
  const res = dbc.products.insertOne({
    shop_id: 'shop_e2e',
    name: 'E2E Test Product',
    slug: 'e2e-test-product',
    description: 'product for e2e flow',
    short_desc: 'e2e',
    price: 250000,
    original_price: 300000,
    cost_price: 150000,
    sku: 'E2E-0001',
    barcode: 'e2e-0001',
    stock: 100,
    sold_count: 0,
    allow_backorder: false,
    thumbnail: '',
    images: [],
    attributes: {},
    tags: ['e2e'],
    weight: 0, length: 0, width: 0, height: 0,
    meta_title: '', meta_description: '',
    status: 'active',
    is_featured: false,
    is_deleted: false,
    average_rating: 0, review_count: 0,
    created_at: new Date(), updated_at: new Date()
  });
  print(res.insertedId.toString());
}
" 2>/dev/null | tr -d '\r' | tail -1)
if [ -z "$PRODUCT_ID" ]; then
  bad "không seed được product vào MongoDB"
  exit 1
fi
ok "product_id=$PRODUCT_ID"

# ------------------------------------------------------------------
EMAIL="e2e_$(date +%s)@example.com"
PASSWORD="StrongPass123!"
PHONE="09$(date +%s | tail -c 9)"
USER_NAME="E2E_User_$(date +%s)"
log "Bước 1 — Đăng ký tài khoản ($EMAIL)..."
REG=$(api POST /api/v1/user/register "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"user_name\":\"$USER_NAME\",\"phone\":\"$PHONE\"}")
if [ "$(code_of "$REG")" = "201" ]; then
  ok "đăng ký thành công (HTTP 201)"
else
  bad "đăng ký thất bại (HTTP $(code_of "$REG")): $(body_of "$REG" | head -c 200)"
  exit 1
fi

# ------------------------------------------------------------------
log "Bước 2 — Lấy OTP từ Redis và xác thực email..."
OTP=$(compose exec -T redis redis-cli --no-auth-warning -a "$REDIS_PASS" GET "email_otp:$EMAIL" 2>/dev/null | tr -d '\r')
if [ -z "$OTP" ]; then
  bad "không đọc được OTP từ Redis (worker chưa kịp lưu?)"
  exit 1
fi
VER=$(api POST /api/v1/user/verify-email-otp "{\"email\":\"$EMAIL\",\"otp\":\"$OTP\"}")
if [ "$(code_of "$VER")" = "200" ]; then
  ok "xác thực OTP thành công (HTTP 200)"
else
  bad "xác thực OTP thất bại (HTTP $(code_of "$VER")): $(body_of "$VER" | head -c 200)"
  exit 1
fi

# ------------------------------------------------------------------
log "Bước 3 — Đăng nhập..."
LOGIN=$(api POST /api/v1/user/login "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
TOKEN=$(json_get "$(body_of "$LOGIN")" '"access_token":"[^"]*"')
if [ "$(code_of "$LOGIN")" = "200" ] && [ -n "$TOKEN" ]; then
  ok "đăng nhập thành công, có access_token"
else
  bad "đăng nhập thất bại (HTTP $(code_of "$LOGIN")): $(body_of "$LOGIN" | head -c 200)"
  exit 1
fi

# ------------------------------------------------------------------
log "Bước 4 — Danh sách sản phẩm..."
PRODS=$(api GET /api/v1/products)
if [ "$(code_of "$PRODS")" = "200" ]; then
  ok "lấy danh sách sản phẩm thành công (HTTP 200)"
else
  bad "lấy sản phẩm thất bại (HTTP $(code_of "$PRODS")): $(body_of "$PRODS" | head -c 200)"
fi

# ------------------------------------------------------------------
log "Bước 5 — Thêm vào giỏ hàng..."
CART=$(api POST /api/v1/cart/items "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":2}" "$TOKEN")
CC=$(code_of "$CART")
if [ "$CC" = "200" ] || [ "$CC" = "201" ]; then
  ok "thêm vào giỏ thành công (HTTP $CC)"
else
  bad "thêm vào giỏ thất bại (HTTP $CC): $(body_of "$CART" | head -c 300)"
fi

# ------------------------------------------------------------------
log "Bước 6 — Xem giỏ hàng..."
CART_GET=$(api GET /api/v1/cart "" "$TOKEN")
if [ "$(code_of "$CART_GET")" = "200" ]; then
  ok "xem giỏ hàng thành công (HTTP 200)"
else
  bad "xem giỏ hàng thất bại (HTTP $(code_of "$CART_GET")): $(body_of "$CART_GET" | head -c 300)"
fi

# ------------------------------------------------------------------
log "Bước 7 — Tạo đơn hàng (checkout, thanh toán COD)..."
ORDER=$(api POST /api/v1/orders "{\"items\":[{\"product_id\":\"$PRODUCT_ID\",\"quantity\":2,\"price\":250000}],\"payment_method\":\"COD\",\"shipping_address\":\"123 Nguyen Trai, Q1, HCM\",\"contact_phone\":\"$PHONE\"}" "$TOKEN")
OC=$(code_of "$ORDER")
ORDER_ID=$(json_get "$(body_of "$ORDER")" '"id":"[^"]*"')
if { [ "$OC" = "200" ] || [ "$OC" = "201" ]; } && [ -n "$ORDER_ID" ]; then
  ok "tạo đơn hàng thành công (HTTP $OC, order_id=$ORDER_ID)"
else
  bad "tạo đơn hàng thất bại (HTTP $OC): $(body_of "$ORDER" | head -c 400)"
  exit 1
fi

# ------------------------------------------------------------------
log "Bước 8 — Xem đơn hàng của tôi..."
MY=$(api GET /api/v1/orders/my "" "$TOKEN")
if [ "$(code_of "$MY")" = "200" ] && echo "$(body_of "$MY")" | grep -q "$ORDER_ID"; then
  ok "đơn hàng xuất hiện trong danh sách /orders/my"
else
  bad "không thấy đơn hàng trong /orders/my (HTTP $(code_of "$MY"))"
fi

# ------------------------------------------------------------------
echo ""
echo "======================================"
echo "E2E TEST RESULT: $PASS passed, $FAIL failed"
echo "======================================"
[ "$FAIL" -eq 0 ] || exit 1
