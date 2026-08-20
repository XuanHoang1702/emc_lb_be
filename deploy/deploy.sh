#!/usr/bin/env bash
#
# CD deploy script — rolling deployment cho cluster EMC LB.
#
# Chiến lược (zero-downtime qua nginx):
#   1. Pull image mới lên server (registry)
#   2. Cập nhật tag image trong docker-compose.yml
#   3. Rolling update từng API node: stop+rm node cũ, up node mới.
#      nginx failover (max_fails=3) tự chuyển request sang node khỏe.
#   4. Smoke test qua nginx sau mỗi node; thất bại -> rollback tag cũ.
#   5. Cập nhật worker (không chặn).
#
# Chạy trên server (workflow SSH vào server rồi execute script này).
#
# Biến môi trường:
#   IMAGE_SERVER / IMAGE_WORKER : image (vd ghcr.io/org/emc_lb-api)
#   IMAGE_TAG / PREVIOUS_TAG    : tag mới / tag cũ (để rollback)
#   COMPOSE_DIR                 : thư mục dự án trên server
#   ENV_FILE                    : tên env file (vd .env.production)
#   GHCR_USERNAME / GHCR_TOKEN  : (tùy chọn) login registry GHCR

set -euo pipefail

IMAGE_SERVER="${IMAGE_SERVER:?IMAGE_SERVER required}"
IMAGE_WORKER="${IMAGE_WORKER:?IMAGE_WORKER required}"
IMAGE_TAG="${IMAGE_TAG:?IMAGE_TAG required}"
COMPOSE_DIR="${COMPOSE_DIR:-/opt/emc_lb}"
# --env-file cho docker compose tính theo cwd; env_file trong compose tính
# theo thư mục chứa compose (docker/). ECC_ENV_FILE thường là ../.env.production.
ENV_FILE="${ECC_ENV_FILE_ABS:-.env.production}"
ECC_ENV_FILE="${ECC_ENV_FILE:-../.env.production}"
COMPOSE_FILE="${COMPOSE_DIR}/docker/docker-compose.yml"

log() { echo -e "\n\033[1;36m[cd]\033[0m $*"; }
die() { echo -e "\033[1;31m[cd] ✘ $*\033[0m"; exit 1; }

compose() {
  # $1 = image tag; chọn image/tag qua biến nội suy của docker-compose.yml.
  ECC_ENV_FILE="$ECC_ENV_FILE" \
  ECC_API_IMAGE="$IMAGE_SERVER" \
  ECC_WORKER_IMAGE="$IMAGE_WORKER" \
  ECC_IMAGE_TAG="$1" \
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "${@:2}"
}

smoke() {
  local i code
  for i in 1 2 3; do
    code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 http://localhost/health || true)
    if [ "$code" = "200" ]; then
      return 0
    fi
    sleep 5
  done
  return 1
}

rollback() {
  log "ROLLBACK về tag $PREVIOUS_TAG..."
  local node
  for node in emc_api_node_1 emc_api_node_2 emc_api_node_3; do
    compose "$PREVIOUS_TAG" up -d --no-deps --no-build "$node" || true
  done
  if smoke; then
    log "Rollback thành công về $PREVIOUS_TAG"
    exit 1  # deploy gốc thất bại, trả code lỗi để workflow đánh dấu fail
  else
    die "Rollback cũng thất bại — cần can thiệp thủ công!"
  fi
}

# ---------------------------------------------------------------
log "CD deploy: $IMAGE_SERVER:$IMAGE_TAG (trước: $PREVIOUS_TAG)"
cd "$COMPOSE_DIR"

# Login registry nếu cung cấp
if [ -n "${GHCR_USERNAME:-}" ] && [ -n "${GHCR_TOKEN:-}" ]; then
  log "Login GHCR..."
  echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USERNAME" --password-stdin
fi

# Pull image mới
log "Pull $IMAGE_SERVER:$IMAGE_TAG và $IMAGE_WORKER:$IMAGE_TAG..."
docker pull "${IMAGE_SERVER}:${IMAGE_TAG}"
docker pull "${IMAGE_WORKER}:${IMAGE_TAG}" || log "  (worker image chưa tồn tại, deploy sau)"

# Rolling update từng node
for node in emc_api_node_1 emc_api_node_2 emc_api_node_3; do
  log "Rolling update $node -> ${IMAGE_TAG}..."
  compose "$IMAGE_TAG" stop "$node" || true
  compose "$IMAGE_TAG" rm -f "$node" || true
  compose "$IMAGE_TAG" up -d --no-deps --no-build "$node" || { rollback; }
  sleep 3
  if ! smoke; then
    log "Smoke test thất bại sau $node — rollback"
    rollback
  fi
  log "$node OK (smoke test qua nginx đạt)"
done

# Worker
log "Update worker -> ${IMAGE_TAG}..."
compose "$IMAGE_TAG" stop emc_worker || true
compose "$IMAGE_TAG" rm -f emc_worker || true
compose "$IMAGE_TAG" up -d --no-deps --no-build emc_worker || log "  (worker không deploy được, kiểm tra thủ công)"

# Smoke test cuối
if smoke; then
  log "✅ Deploy hoàn tất — $IMAGE_TAG đang phục vụ qua nginx"
else
  die "Smoke test cuối thất bại"
fi