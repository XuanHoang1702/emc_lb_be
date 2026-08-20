// ECC / EMC LB — Load test qua nginx load balancer
//
// Mô phỏng luồng mua hàng thực tế: health -> browse -> detail -> register ->
// login -> add to cart -> checkout. Chạy song song nhiều VU như giờ Black Friday.
//
// Cách chạy:
//   k6 run loadtest/checkout-flow.js
//   k6 run --only smoke loadtest/checkout-flow.js
//   k6 run --only soak  loadtest/checkout-flow.js
//   k6 run --only spike loadtest/checkout-flow.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

// Danh sách sản phẩm giả định để thêm vào giỏ. Trong CI, seed script sẽ chèn
// các product_id thật vào biến môi trường PRODUCT_IDS (phân tách bằng dấu phẩy).
const FALLBACK_PRODUCTS = ['prod_1', 'prod_2', 'prod_3'];
const PRODUCTS = (__ENV.PRODUCT_IDS || '').split(',').filter(Boolean);
const productPool = PRODUCTS.length > 0 ? PRODUCTS : FALLBACK_PRODUCTS;

// Số user đã seed + verify email (src/cmd/loadtest-seed), dùng để login.
// Mặc định 10; CI seed 200 user nên đủ cho spike 2000 VU xoay vòng.
const LOADTEST_USER_COUNT = parseInt(__ENV.LOADTEST_USER_COUNT || '10', 10) || 10;

// Ngưỡng SLO (Service Level Objective)
const SLO = {
  smoke: {
    p95: 1000, // ms
    errorRate: 0, // %
    executors: [{
      exec: 'smoke',
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
    }],
    thresholds: {
      http_req_failed: ['rate<0.01'],
      http_req_duration: ['p(95)<1000'],
      checks: ['rate>0.99'],
    },
  },
  rampUp: {
    p95: 800,
    errorRate: 1,
    executors: [{
      exec: 'buyer',
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 100 },
        { duration: '2m', target: 300 },
        { duration: '2m', target: 500 },
        { duration: '1m', target: 0 },
      ],
    }],
    thresholds: {
      http_req_failed: ['rate<0.01'],
      http_req_duration: ['p(95)<800'],
      checks: ['rate>0.95'],
    },
  },
  soak: {
    p95: 800,
    errorRate: 1,
    executors: [{
      exec: 'buyer',
      executor: 'constant-vus',
      vus: 100,
      duration: '10m',
    }],
    thresholds: {
      http_req_failed: ['rate<0.01'],
      http_req_duration: ['p(95)<800'],
      checks: ['rate>0.95'],
    },
  },
  spike: {
    p95: 1500,
    errorRate: 5,
    executors: [{
      exec: 'buyer',
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 2000 },
        { duration: '30s', target: 2000 },
        { duration: '1m', target: 0 },
      ],
    }],
    thresholds: {
      http_req_failed: ['rate<0.05'],
      http_req_duration: ['p(95)<1500'],
      checks: ['rate>0.90'],
    },
  },
};

// Mặc định chạy toàn bộ scenario, trừ khi người dùng chỉ định --only
const ONLY = __ENV.ONLY || __ENV.K6_OPTION_ONLY || null;
const scenarioMap = {
  smoke: 'smoke',
  'ramp-up': 'rampUp',
  rampup: 'rampUp',
  soak: 'soak',
  spike: 'spike',
};

let scenarios = {};
let thresholds = {};
let executorsToRun = ONLY ? [scenarioMap[ONLY]] : ['smoke', 'rampUp', 'soak', 'spike'];
executorsToRun = executorsToRun.filter(Boolean);

for (const name of executorsToRun) {
  const cfg = SLO[name];
  scenarios[name] = cfg.executors[0];
  Object.assign(thresholds, cfg.thresholds);
}

export const options = {
  scenarios,
  thresholds,
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'max'],
};

function apiKeyHeader() {
  return __ENV.API_KEY ? { 'x-api-key': __ENV.API_KEY } : {};
}

// ---- Các request nghiệp vụ -------------------------------------------------

function healthCheck() {
  const res = http.get(`${BASE_URL}/health`, {
    tags: { name: 'health' },
  });
  check(res, {
    'health: 200 + up': (r) => r.status === 200 && JSON.parse(r.body).status === 'up',
  });
}

function listProducts() {
  const res = http.get(`${BASE_URL}/api/v1/products?limit=20`, {
    tags: { name: 'list_products' },
  });
  check(res, {
    'products list: 200': (r) => r.status === 200,
  });
  try {
    const body = JSON.parse(res.body);
    const products = body.data || [];
    if (Array.isArray(products) && products.length > 0) {
      return products;
    }
  } catch (e) {
    // response không parse được — dùng pool mặc định
  }
  return null;
}

function getProductDetail(id) {
  const res = http.get(`${BASE_URL}/api/v1/products/${id}`, {
    tags: { name: 'product_detail' },
  });
  check(res, {
    'product detail: 200': (r) => r.status === 200,
  });
}

function loginUser(email, password) {
  const payload = JSON.stringify({
    email,
    password,
  });
  const res = http.post(`${BASE_URL}/api/v1/user/login`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'login' },
  });
  check(res, {
    'login: 200': (r) => r.status === 200,
  });
  try {
    const body = JSON.parse(res.body);
    return body.data && body.data.access_token ? body.data.access_token : null;
  } catch (e) {
    return null;
  }
}

function addToCart(token, productId) {
  const payload = JSON.stringify({
    product_id: productId,
    quantity: randomIntBetween(1, 3),
  });
  const res = http.post(`${BASE_URL}/api/v1/cart/items`, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    tags: { name: 'add_to_cart' },
  });
  check(res, {
    'add to cart: 2xx': (r) => r.status >= 200 && r.status < 300,
  });
  return res;
}

function checkout(token, productId) {
  const payload = JSON.stringify({
    items: [{ product_id: productId, quantity: 1 }],
    payment_method: 'COD',
    shipping_address: 'Hanoi, Vietnam',
    contact_phone: '0900000000',
  });
  const res = http.post(`${BASE_URL}/api/v1/orders`, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    tags: { name: 'checkout' },
  });
  check(res, {
    'checkout: 2xx': (r) => r.status >= 200 && r.status < 300,
  });
  return res;
}

// ---- Executors --------------------------------------------------------------

// Dùng user đã được seed + xác thực email sẵn (src/cmd/loadtest-seed) để
// login thành công. VU và iteration ánh xạ sang user trong pool.
function pickUser() {
  const i = (__VU * 1000 + __ITER) % LOADTEST_USER_COUNT;
  const email = `loadtest_${String(i + 1).padStart(3, "0")}@example.com`;
  return { email, password: 'LoadTest123!' };
}

export function smoke() {
  healthCheck();
  const products = listProducts();
  const pool = products && products.length > 0 ? products.map((p) => p.id || p) : productPool;
  const productId = randomItem(pool);

  getProductDetail(productId);
  const token = loginUser(pickUser().email, pickUser().password);
  if (token) {
    addToCart(token, productId);
    checkout(token, productId);
  }
  sleep(randomIntBetween(1, 3));
}

export function buyer() {
  const productId = randomItem(productPool);
  // Xác suất: 20% health, 80% luồng mua hàng — đợt spike thường có nhiều
  // request đọc hơn là ghi.
  const roll = Math.random();
  if (roll < 0.2) {
    healthCheck();
    listProducts();
    sleep(randomIntBetween(0.5, 2));
    return;
  }

  getProductDetail(productId);
  const token = loginUser(pickUser().email, pickUser().password);
  if (token) {
    addToCart(token, productId);
    if (Math.random() < 0.7) {
      checkout(token, productId);
    }
  }
  sleep(randomIntBetween(0.5, 3));
}

// Khi chạy full (không --only), name executors phải khớp với exported function.
export default function () {
  const exec = __ENV.K6_EXEC_NAME || '';
  if (exec === 'smoke') smoke();
  else if (exec === 'rampUp' || exec === 'spike') buyer();
  else buyer();
}