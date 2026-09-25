# EMC LB Performance Engineering Report

## Overview
This document contains the performance engineering test artifacts and the final report template for the EMC LB project. The tests are designed to be executed in a safe, isolated Docker environment using `k6`.

## Directory Structure
- `scripts/`: Contains the k6 load test scenarios.
- `utils/`: Helper scripts for k6.
- `data/`: Contains `seed.json` with generated test data (users, products, categories).
- `run.sh`: Helper script to execute k6 tests in Docker against the local environment.

## Running the Tests
1. Ensure the local Docker environment is running: `make deploy` or `docker compose up -d`.
2. Wait for services to be healthy (`docker ps`).
3. Seed the database (instructions below).
4. Run a test using the wrapper script: `./run.sh scripts/01_baseline.js`

---

# Final Performance Report (Sandbox Environment)

## 1. Environment
- **CPU**: 12 Cores (Sandbox)
- **RAM**: 16 GB (Sandbox)
- **Docker configuration**: `docker-compose.yml` with limits (API nodes: 0.5 CPU, 256MB RAM)
- **API node count**: 3 nodes behind NGINX
- **DB versions**: PostgreSQL 15, MongoDB 7, Redis 7
- **Worker configuration**: Asynq Worker (1 container, default queues)
- **k6 version**: grafana/k6 Docker Image

## 2. Test Scenarios
- `01_baseline.js`: Tests `/health` and basic `/api/v1/categories`.
- `02_public_read.js`: Tests paginated product list reads `/api/v1/products`.
- `03_product_detail.js`: Tests concurrent reads of a single hot product.
- `06_auth.js`: Tests login `/api/v1/users/login`.
- `08_order_creation.js`: Tests concurrent order creation for limited stock.

## 3. Results (To be filled from actual k6 runs)

| Scenario | Load (VUs) | RPS | p50 | p95 | p99 | Error Rate |
|---|---:|---:|---:|---:|---:|---:|
| 01 Baseline | 50 | TBA | TBA | TBA | TBA | TBA |
| 02 Public Read | 100 | TBA | TBA | TBA | TBA | TBA |
| 03 Product Detail | 100 | TBA | TBA | TBA | TBA | TBA |
| 06 Auth | 50 | TBA | TBA | TBA | TBA | TBA |
| 08 Order Creation | 100 | TBA | TBA | TBA | TBA | TBA |

*(Note: Actual execution numbers depend on the hardware where `run.sh` is executed. The sandbox container may throttle early.)*

## 4. Cache Results
- **Expected**: Product Detail and Category endpoints will heavily hit Redis.
- **Hit Ratio**: Monitor via `redis_exporter` Grafana dashboard.

## 5. Database Results
- **PostgreSQL**: Connections limited to 10 max open per API node (total 30).
- **MongoDB**: Connection pool max size 50 per node (total 150).

## 6. Redis Results
- Used for: Product Cache, Inventory Lua scripts, Asynq tasks.
- **Connection Pool**: 20 per API node (total 60).

## 7. Worker Results
- Task queues: `critical`, `default`.
- Concurrency bounded by Asynq worker configuration.

## 8. Multi-node Results
- Traffic is round-robined by NGINX (`emc_api_node_1`, `emc_api_node_2`, `emc_api_node_3`).

## 9. Failure Tests
*(Simulate failures by running `docker stop emc_lb_api_node_1` during a test and observe error rates in k6).*

## 10. Bottleneck
- **Theoretical primary bottleneck**: API CPU due to 0.5 CPU limit per container.
- **Theoretical secondary bottleneck**: Redis single-threaded execution for Lua inventory script if RPS exceeds 50k (unlikely to reach in sandbox).

## 11. Optimizations
*(To be applied based on empirical test data from the user's local workstation)*

## 12. Remaining Bottlenecks
*(To be populated post-testing)*

## 13. Recommended Next Engineering Task
*(To be determined after full-scale production-like load tests)*
