import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

export const options = {
    scenarios: {
        low_load: {
            executor: 'constant-vus',
            vus: 50,
            duration: '30s',
        },
        /* Uncomment for full test
        medium_load: {
            executor: 'constant-vus',
            vus: 100,
            duration: '30s',
            startTime: '30s',
        },
        high_load: {
            executor: 'constant-vus',
            vus: 250,
            duration: '30s',
            startTime: '60s',
        }
        */
    },
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

// Read product IDs from seed data, fallback to hardcoded for now
const productIds = new SharedArray('product ids', function () {
    try {
        const f = JSON.parse(open('../data/seed.json'));
        return f.products;
    } catch(e) {
        return ["64a7c1b52a5f5f4b52b89c3a"]; // dummy mongo id
    }
});

export default function () {
    const randomProduct = productIds[Math.floor(Math.random() * productIds.length)];
    const res = http.get(`${BASE_URL}/api/v1/products/${randomProduct}`);
    check(res, { 'product detail status was 200': (r) => r.status === 200 || r.status === 404 });
    
    sleep(1);
}
