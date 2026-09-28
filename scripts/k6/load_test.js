import http from 'k6/http';
import { check, sleep } from 'k6';

// The k6 load test configuration
export const options = {
    scenarios: {
        mixed_load: {
            executor: 'ramping-arrival-rate',
            startRate: 10,
            timeUnit: '1s',
            preAllocatedVUs: 50,
            maxVUs: 500,
            stages: [
                { duration: '30s', target: 10 },  // Warm up
                { duration: '1m', target: 50 },   // Normal load
                { duration: '1m', target: 100 },  // High load
                { duration: '30s', target: 200 }, // Stress
                { duration: '10s', target: 0 },   // Cooldown
            ],
        },
    },
    thresholds: {
        // We expect less than 1% of requests to fail
        http_req_failed: ['rate<0.01'],
        // 95% of requests should complete within 500ms
        http_req_duration: ['p(95)<500'],
    },
};

// Use environment variable or default to localhost
// When running via docker compose, it might be 'http://nginx:80'
const BASE_URL = __ENV.API_URL || 'http://localhost';

export default function () {
    // Randomly choose an endpoint to hit
    // 30% of traffic goes to /health, 70% goes to /api/v1/products
    const rand = Math.random();

    if (rand < 0.3) {
        // Lightweight endpoint test
        const res = http.get(`${BASE_URL}/health`);
        check(res, {
            'health status is 200': (r) => r.status === 200,
        });
    } else {
        // Read-heavy endpoint test
        const res = http.get(`${BASE_URL}/api/v1/products`);
        check(res, {
            'products status is 200': (r) => r.status === 200,
        });
    }

    // Sleep for a short duration to simulate real user wait time
    // Note: In arrival-rate executor, sleep doesn't affect the request rate,
    // but it holds the VU for this duration.
    sleep(0.5);
}
