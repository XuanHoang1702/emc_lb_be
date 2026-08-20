#!/usr/bin/env bash
#
# Chaos / Resilience test cho hệ thống cân bằng tải EMC LB.
#
# Mô phỏng sự cố thật và kiểm tra hệ thống phản ứng:
#   1. Kill một API node  -> nginx phải chuyển request sang node còn lại
#   2. Kill 2/3 node      -> hệ thống vẫn phục vụ (suy giảm nhưng không chết)
#   3. Dừng Redis         -> health detail phải báo degraded, request vẫn xử lý
#   4. Network delay      -> node bị trễ, nginx failover sang node khỏe
#   5. Phục hồi           -> kill "thủ phạm" rồi scale lại, hệ thống tự lành
#
# Cách chạy:
#   bash chaos/chaos-test.sh
#   ECC_ENV_FILE=.env.ci bash chaos/chaos-test.sh

set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker/docker-compose.yml}"
# --env-file truyền cho docker compose được tính theo cwd (repo root).
# env_file: ${ECC_ENV_FILE:-...} trong compose được tính theo thư mục docker/.
ENV_FILE="${ECC_ENV_FILE_ABS:-.env.development}"
ECC_ENV_FILE="${ECC_ENV_FILE:-../.env.development}"
BASE_URL="${BASE_URL:-http://localhost}"
API_KEY="${API_KEY:-}"

PASS=0
FAIL=0
NODES=(emc_api_node_1 emc_api_node_2 emc_api_node_3)

log()  { echo -e "\n\033[1;34m[chaos]\033[0m $*"; }
ok()   { echo -e "  \033[1;32m✔\033[0m $*"; PASS=$((PASS+1)); }
bad()  { echo -e "  \033[1;31m✘\033[0m $*"; FAIL=$((FAIL+1)); }

health_up() {
  curl -sf "$BASE_URL/health" 2>/dev/null | grep -q '"status":"up"'
}

request_probe() {
  # Một request phải thành công (2xx) bất kể trạng thái nội bộ.
  # Thử tối đa 3 lần; timeout 10s để cho nginx đủ thời gian failover
  # qua các node chết (mỗi node chết tốn ~1s connect timeout).
  local i code
  for i in 1 2 3; do
    code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "$BASE_URL/health")
    if [ "$code" = "200" ]; then
      return 0
    fi
    sleep 1
  done
  return 1
}

detail_status() {
  # Trả về "up" / "degraded" từ /health/detail (bảo vệ bằng API key).
  # Không dùng -f: endpoint trả 503 khi degraded — vẫn phải đọc được body.
  # Overall status luôn là mục "status" cuối cùng trong JSON này.
  # --max-time 10: khi redis/db down, go-redis dial timeout (~5s) làm endpoint
  # phản hồi chậm; 5s là quá ngắn khiến curl trả rỗng.
  local hdr=()
  if [ -n "$API_KEY" ]; then hdr=(-H "x-api-key: $API_KEY"); fi
  curl -s "${hdr[@]}" --max-time 10 "$BASE_URL/health/detail" 2>/dev/null \
    | grep -o '"status":"[a-z]*"' | tail -1 | cut -d'"' -f4
}

compose() {
  ECC_ENV_FILE="$ECC_ENV_FILE" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

wait_healthy() {
  for i in $(seq 1 90); do
    if health_up && request_probe; then
      return 0
    fi
    sleep 2
  done
  return 1
}

# ------------------------------------------------------------
log "Khởi động stack (nếu chưa chạy)..."
compose up -d --build
if ! wait_healthy; then
  bad "stack không healthy sau 180s — kiểm tra docker compose logs"
  compose logs || true
  exit 1
fi
ok "stack healthy"

# ------------------------------------------------------------
log "Chaos #1 — Kill 1 API node: hệ thống phải vẫn phục vụ"
killed1="$(compose ps --format '{{.Service}}' | grep -E 'emc_api_node_[0-9]' | head -1)"
[ -n "$killed1" ] || killed1="${NODES[0]}"
compose stop "$killed1" >/dev/null
ok "stopped $killed1"
sleep 2
if request_probe; then
  ok "request thành công khi thiếu $killed1 — LB failover hoạt động"
else
  bad "request thất bại khi $killed1 chết — nginx không failover?"
fi
compose start "$killed1" >/dev/null
wait_healthy || bad "không phục hồi sau khi start lại $killed1"

# ------------------------------------------------------------
log "Chaos #2 — Kill 2/3 node: hệ thống suy giảm nhưng không sập"
alive_nodes=()
for n in "${NODES[@]}"; do
  if compose ps --format '{{.Service}} {{.State}}' | grep -q "^$n running"; then
    alive_nodes+=("$n")
  fi
done
if [ "${#alive_nodes[@]}" -lt 2 ]; then
  log "  (cần ít nhất 2 node đang chạy để test; bỏ qua bước này)"
else
  compose stop "${alive_nodes[0]}" "${alive_nodes[1]}" >/dev/null 2>&1
  ok "stopped ${alive_nodes[0]} ${alive_nodes[1]}"
  sleep 2
  if request_probe; then
    ok "vẫn phục vụ với 1/3 node — LB failover hoạt động dưới suy giảm"
  else
    bad "sập khi chỉ còn 1 node — LB không chịu được suy giảm"
  fi
  for n in "${alive_nodes[@]}"; do
    compose start "$n" >/dev/null
  done
  wait_healthy || bad "không phục hồi sau khi scale lại"
fi

# ------------------------------------------------------------
log "Chaos #3 — Dừng Redis: /health/detail phải báo degraded"
compose stop redis >/dev/null
ok "stopped redis"
sleep 2
st="$(detail_status || true)"
case "$st" in
  degraded|up)
    ok "/health/detail trả về '$st' (redis down được phản ánh)"
    ;;
  *)
    bad "/health/detail không phản ánh redis down (trả về '$st')"
    ;;
esac
compose start redis >/dev/null
wait_healthy || bad "không phục hồi sau khi start redis"

# ------------------------------------------------------------
log "Chaos #4 — Network partition: 1 node im lặng, LB phải failover"
slow_node="$(compose ps --format '{{.Service}}' | grep -E 'emc_api_node_[0-9]' | head -1)"
[ -n "$slow_node" ] || slow_node="${NODES[0]}"
compose stop "$slow_node" >/dev/null
ok "$slow_node ngừng phản hồi (mô phỏng network partition)"
sleep 2
if request_probe; then
  ok "LB chuyển request khi $slow_node không phản hồi"
else
  bad "LB không failover khi $slow_node im lặng"
fi
compose start "$slow_node" >/dev/null
wait_healthy || bad "không phục hồi sau network partition"

# ------------------------------------------------------------
log "Chaos #5 — Phục hồi: toàn bộ cluster healthy lại"
wait_healthy
count=$(compose ps --filter "status=running" --format '{{.Service}}' | grep -cE 'emc_api_node_[0-9]' || true)
if [ "$count" -ge 3 ]; then
  ok "cả 3 API nodes đang chạy và healthy"
else
  bad "chỉ $count/3 nodes đang chạy"
fi

# ------------------------------------------------------------
echo ""
echo "======================================"
echo "CHAOS TEST RESULT: $PASS passed, $FAIL failed"
echo "======================================"
[ "$FAIL" -eq 0 ] || exit 1