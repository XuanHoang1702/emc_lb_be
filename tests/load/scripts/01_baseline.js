import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

export default function () {
    const res1 = http.get(`${BASE_URL}/health`);
    check(res1, { 'health status was 200': (r) => r.status === 200 });

    const res2 = http.get(`${BASE_URL}/api/v1/categories`);
    check(res2, { 'categories status was 200': (r) => r.status === 200 });
    
    sleep(1);
}
