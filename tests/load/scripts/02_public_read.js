import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

export default function () {
    // Test product listing which might be heavy
    const res1 = http.get(`${BASE_URL}/api/v1/products?page=1&limit=20`);
    check(res1, { 'products list status was 200': (r) => r.status === 200 });
    
    sleep(1);
}
