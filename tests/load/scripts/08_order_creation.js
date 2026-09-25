import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        concurrent_orders: {
            executor: 'constant-vus',
            vus: 100, // 100 concurrent attempts
            duration: '30s',
        }
    },
    thresholds: {
        http_req_duration: ['p(95)<1000'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

export default function () {
    const payload = JSON.stringify({
        product_id: "64a7c1b52a5f5f4b52b89c3a", // hot product id
        quantity: 1,
        payment_method: "sepay"
    });
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer DUMMY_TOKEN'
        },
    };

    const res = http.post(`${BASE_URL}/api/v1/orders`, payload, params);
    check(res, { 
        'order status was 200 or 400 (stock empty)': (r) => r.status === 200 || r.status === 400 
    });
    
    sleep(0.5);
}
